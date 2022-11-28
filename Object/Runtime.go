package Object

import (
	"WebDav-ClientWeb/Object/Config"
	"WebDav-ClientWeb/Object/Config/Log"
	"bytes"
	"context"
	"fmt"
	"github.com/klarkxy/gohtml"
	"github.com/studio-b12/gowebdav"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type HandlerServer struct {
	ServerInfo     Config.Server
	webdavFilePath []string
}

//生成http服务对象
func MakeS3HttpServer(config Config.Yaml) {
	var serverList []Config.Server
	var serverObjectList []*http.Server
	for _, list := range config.ServerList {
		if list.Enable {
			serverList = append(serverList, list)
		}
	}
	for _, server := range serverList {
		serverhandler := HandlerServer{}
		serverhandler.ServerInfo = server
		webserver := http.Server{
			Addr:    server.ListenPort,
			Handler: serverhandler,
		}
		serverObjectList = append(serverObjectList, &webserver)
	}
	Log.AppLog.Infoln("Server config load Ok!")
	RunHttpServer(serverList, serverObjectList)
}

//运行http服务
func RunHttpServer(serverlist []Config.Server, httpSrv []*http.Server) {
	for i, serverObject := range httpSrv {
		Log.AppLog.Infoln(serverlist[i].Name + " server listening at: " + serverlist[i].ListenPort + "/" + serverlist[i].Name)
		if serverlist[i].Options.UseTLS.Enable {
			go func() {
				err := serverObject.ListenAndServeTLS(serverlist[i].Options.UseTLS.CertFile, serverlist[i].Options.UseTLS.CertKey)
				if err != nil {
					Log.AppLog.Warningln(err.Error())
				}
			}() //监听https服务
		} else {
			go func() {
				err := serverObject.ListenAndServe()
				if err != nil {
					Log.AppLog.Warningln(err.Error())
				}
			}() //协程并发监听http服务
		}
	}
	Log.AppLog.Infoln("The http service is started and the program is started!")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	//程序堵塞
	case <-sigs: //检测Ctrl+c退出程序命令
		for i, serverObject := range httpSrv {
			Log.AppLog.Infoln("Shutting down " + serverlist[i].Name + " instance gracefully...")
			er := serverObject.Shutdown(context.Background()) //平滑关闭Http Server线程
			if er != nil {
				Log.SetReportCaller(true)
				Log.AppLog.Fatalln(er.Error())
				Log.SetReportCaller(false)
			}
			Log.AppLog.Infoln("Instance " + serverlist[i].Name + " has exited safely!")
		}
	}
}

//处理请求
func (webserver HandlerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url, _ := url.QueryUnescape(r.URL.RequestURI())
	urlArray := strings.Split(url, "/")
	//获取IP地址
	clientIP := GetIP(r)
	//判断url是哪一种请求
	//网站图标
	Request := clientIP + " --> " + r.Host + "," + r.Proto + "," + r.UserAgent() + "," + r.RequestURI
	if r.URL.RequestURI() == "/favicon.ico" {
		w.Header().Set("Location", webserver.ServerInfo.Options.Favicon)
		w.WriteHeader(301)
		Log.AppLog.Infoln(Request, "Code:", 301)
		return
	}
	if urlArray[1] != "d" && urlArray[1] != webserver.ServerInfo.Name {
		fmt.Fprintln(w, ErrorPage_404(webserver.ServerInfo.Name))
		Log.AppLog.Infoln(Request, "Code:", 404)
		return
	}
	if urlArray[1] == "d" {
		if len(urlArray) <= 3 && len(urlArray) >= 2 {
			fmt.Fprintln(w, ErrorPage_404(webserver.ServerInfo.Name))
			Log.AppLog.Infoln(Request, "Code:", 404)
			return
		} else if len(urlArray) == 4 && urlArray[3] == "" {
			fmt.Fprintln(w, ErrorPage_404(webserver.ServerInfo.Name))
			Log.AppLog.Infoln(Request, "Code:", 404)
			return
		}
	}

	//文件下载
	if len(urlArray) >= 3 && strings.EqualFold(urlArray[1], "d") {
		WebdavClient := gowebdav.NewClient(webserver.ServerInfo.Host, webserver.ServerInfo.UserName, webserver.ServerInfo.PassWD)
		er := WebdavClient.Connect()
		if er != nil {
			Log.AppLog.Errorln(er.Error())
			fmt.Fprintln(w, er.Error())
			return
		}
		str := strings.SplitN(url, "/", 4)
		url := str[3]
		objectStream, er := WebdavClient.ReadStream(url)
		defer objectStream.Close()
		bufnew := new(bytes.Buffer)
		_, err := bufnew.ReadFrom(objectStream)
		if err != nil {
			Log.AppLog.Errorln(err.Error())
			fmt.Fprintln(w, err.Error())
			return
		}
		reader := bytes.NewReader(bufnew.Bytes())
		ContentTyper := http.DetectContentType(bufnew.Bytes())
		// 资源关闭
		if err != nil {
			Log.AppLog.Warningln("sendFile1", err.Error())
			http.NotFound(w, r)
			return
		}
		w.Header().Add("Accept-ranges", "bytes")
		w.Header().Add("Content-Disposition", "attachment; filename="+urlArray[len(urlArray)-1])
		w.Header().Add("Access-Control-Allow-Origin", webserver.ServerInfo.Options.AccessControlAllowOrigin)
		w.Header().Add("content-type", ContentTyper)
		var start, end int64
		//fmt.Println(request.Header,"\n")
		if ra := r.Header.Get("Range"); ra != "" {
			if strings.Contains(ra, "bytes=") && strings.Contains(ra, "-") {
				fmt.Sscanf(ra, "bytes=%d-%d", &start, &end)
				if end == 0 {
					end = reader.Size() - 1
				}
				if start > end || start < 0 || end < 0 || end >= reader.Size() {
					w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
					Log.AppLog.Warningln("sendFile2 start:", start, "end:", end, "size:", reader.Size())
					return
				}
				w.Header().Add("Content-Length", strconv.FormatInt(end-start+1, 10))
				w.Header().Add("Content-Range", fmt.Sprintf("bytes %v-%v/%v", start, end, reader.Size()))
				w.WriteHeader(http.StatusPartialContent)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			// 非断点续传
			Log.AppLog.Infoln(Request, "Code", 200)
			w.Header().Add("Content-Length", strconv.FormatInt(reader.Size(), 10))
			start = 0
			end = reader.Size() - 1
		}
		_, err = reader.Seek(start, 0)
		// add compare
		if start == (end - start + 1) {
			return
		}
		if err != nil {
			Log.AppLog.Warningln("SentFile3:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		n := 512
		buf := make([]byte, n)
		for {
			if end-start+1 < int64(n) {
				n = int(end - start + 1)
			}
			//原生 io
			_, er := io.CopyBuffer(w, reader, buf)
			if er != nil {
				Log.AppLog.Errorln(er.Error())
				return
			}
			start += int64(n)
			if start >= end+1 {
				return
			}
		}
	}
	//显示首页
	str := strings.SplitN(url, "/", 3)
	webserver.webdavFilePath = str
	html := HomePage(webserver)
	fmt.Fprintln(w, html)
	Log.AppLog.Infoln(Request, "Code", 200)
}

//生成404页面
func ErrorPage_404(url string) string {
	html := gohtml.NewHtml()
	html.Head().Title().Text("Url Error")
	html.Meta().Charset("utf-8")
	html.Body().H3().Text("404 链接不正确，请访问正确的地址")
	html.Body().A().Href("/" + url).Text(url)
	return html.String()
}

//生成主页
func HomePage(webserver HandlerServer) string {
	WebdavClient := gowebdav.NewClient(webserver.ServerInfo.Host, webserver.ServerInfo.UserName, webserver.ServerInfo.PassWD)
	er := WebdavClient.Connect()
	if er != nil {
		Log.AppLog.Error(er.Error())
		return "Create Client:" + er.Error()
	}
	var objectList []os.FileInfo
	if len(webserver.webdavFilePath) == 2 {
		objectList, er = WebdavClient.ReadDir("")
	} else {
		objectList, er = WebdavClient.ReadDir(webserver.webdavFilePath[2])
	}
	if er != nil {
		Log.AppLog.Errorln(er.Error())
		return er.Error()
	}
	//var i = 0 //对象计数
	var InfoList []Config.ObjectInfo
	for _, object := range objectList {
		//i++
		Info := Config.ObjectInfo{}
		//Info.Num = i
		Info.Key = object.Name()
		Info.Size = object.Size()
		Info.LsatModified = object.ModTime()
		Info.IsDir = object.IsDir()
		InfoList = append(InfoList, Info)
	}
	return makeHomePageHtml(InfoList, webserver.ServerInfo, webserver.webdavFilePath)
}

func makeBreadCrumb(divframe *gohtml.GoTag, urlpath []string, servername string) (*gohtml.GoTag, string) {
	nav := divframe.Body().Div().Nav().Attr("aria-label", "breadcrumb")
	ol := nav.Body().Ol().Class("breadcrumb")
	url := "/" + servername
	if len(urlpath) == 2 && urlpath[1] == servername {
		ol.Body().Li().Class("breadcrumb-item active").Attr("aria-current", "page").Text("Home")
		return divframe, url
	}
	ol.Body().Li().Class("breadcrumb-item").Body().A().Text("Home").Href(url)
	strpath := strings.Split(urlpath[2], "/")
	for i, str := range strpath {
		url += "/" + str
		if len(strpath)-1 == i {
			ol.Body().Li().Class("breadcrumb-item active").Attr("aria-current", "page").Text(str)
		} else {
			ol.Body().Li().Class("breadcrumb-item").Body().A().Text(str).Href(url)
		}
	}
	return divframe, url
}

//生成主页html对象
func makeHomePageHtml(infolist []Config.ObjectInfo, serverInfo Config.Server, currentUrlPath []string) string {
	//html构建
	bootstrap := gohtml.NewHtml()
	bootstrap.Html().Lang("zh-CN")
	// Meta部分
	bootstrap.Meta().Charset("utf-8")
	bootstrap.Meta().Http_equiv("X-UA-Compatible").Content("IE=edge")
	bootstrap.Meta().Name("viewport").Content("width=device-width, initial-scale=1")
	// Css引入
	bootstrap.Link().Href("https://cdnjs.cloudflare.com/ajax/libs/bootstrap/5.2.0/css/bootstrap.min.css").Rel("stylesheet")
	// Js引入
	bootstrap.Script().Src("https://cdnjs.cloudflare.com/ajax/libs/jquery/3.6.0/jquery.min.js")
	bootstrap.Script().Src("https://cdnjs.cloudflare.com/ajax/libs/bootstrap/5.2.0/js/bootstrap.min.js")
	// Head
	bootstrap.Head().Title().Text("WebDav Server " + serverInfo.Name)
	// Body
	divframe := bootstrap.Body().Div()                                                                         //声明容器
	divframe.Class("container-md")                                                                             //容器样式
	divframe.H1().Text("WebDav - WebDav Client Web Site").Class("container text-center")                       //大标题
	divframe.Body().Hr()                                                                                       //分割线
	divframe.Body().Div().Class("alert alert-success").Text("The current WebDav Server is " + serverInfo.Name) //桶标识
	var url string
	divframe, url = makeBreadCrumb(divframe, currentUrlPath, serverInfo.Name)
	table := divframe.Body().Div().Class("rounded border border-success p-2 mb-2").Table().Class("table border-primary table-hover")
	tr := table.Body().Thead().Class("table-light").Tr()
	tr.Body().Th().Attr("scope", "col").Text("#")
	tr.Body().Th().Attr("scope", "col").Text("Object")
	tr.Body().Th().Attr("scope", "col").Text("Info")
	tb := table.Body().Tbody().Class("table-group-divider")
	for i, object := range infolist {
		tr := tb.Body().Tr()
		if object.IsDir {
			tr.Body().Th().Attr("scope", "row").Span().Class("badge rounded-pill text-bg-light").Text(strconv.Itoa(i + 1))
			td1 := tr.Body().Td()
			td1.Span().Class("badge text-bg-warning").Text("Dir")
			td1.A().Href(url + "/" + object.Key).Target("_black").Text(object.Key)

			td2 := tr.Body().Td()
			td2.Body().Span().Class("badge text-bg-warning").Text(object.LsatModified.Format(time.ANSIC))
		} else {
			tr.Body().Th().Attr("scope", "row").Span().Class("badge rounded-pill text-bg-light").Text(strconv.Itoa(i + 1))
			td1 := tr.Body().Td()
			td1.Span().Class("badge text-bg-info").Text("File")
			td1.A().Href("/d" + url + "/" + object.Key).Target("_black").Text(object.Key)
			td2 := tr.Body().Td()
			td2.Body().Span().Class("badge text-bg-primary").Text(getObjectSizeSuitableUnit(object.Size))
			td2.Body().Span().Class("badge text-bg-warning").Text(object.LsatModified.Format(time.ANSIC))
		}

	}
	divframe.Body().Hr()
	footerdiv := divframe.Body().Div().Class("container-sm text-center").Div().Class("row justify-content-sm-center").Div().Class("col-md-6")
	ul := footerdiv.Body().Ul()
	if serverInfo.Options.BeianMiit != "" {
		ul.Class("list-group list-group-horizontal")
	} else {
		ul.Class("list-group")
	}
	leftli := ul.Body().A().Class("list-group-item list-group-item-action list-group-item-light").Href("https://github.com/xinjiajuan/WebDav-WebClient_go").Target("_black").Text("Powered by")
	leftli.Body().Span().Class("badge rounded-pill text-bg-success").Text("WebDav Client Web " + Config.Version)
	if serverInfo.Options.BeianMiit != "" {
		ul.Body().A().Class("list-group-item list-group-item-action list-group-item-light").Href("https://beian.miit.gov.cn/").Target("_black").Span().Class("badge text-bg-danger").Text(serverInfo.Options.BeianMiit)
	}
	divframe.Body().Br()
	return bootstrap.String()
}
