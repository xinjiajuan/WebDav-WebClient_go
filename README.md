# WebDav-ClientWeb
把webdav服务变成文件下载站
## 使用
```shell
# Run
./WebDav-ClientWeb -c config.yaml 
# Help
./WebDav-ClientWeb -h
```
## 配置
`config.yaml`
```yaml
server:
    # you can make server in array
  - Webdavname: servername
    listenPort: '[]:8080' #string ,ip or port,support ipv4 |ipv6
    enable: true # enable is listen server?
    webdavhost: 'https://xyz.domain.com/dav' #webdav host
    user: 'xxxxxxxx@domain.com' #webdav user
    passwd: o2g8**********************  #webdav password
    weboptions: #web sub option
      useTLS: # ssl sub option
        enable: false # enable is web ssl?
        certFile: ""  # cert path
        certKey: "" #certkey path
      access-control-allow-origin: "*" #资源跨域策略,只对下载链接有效,主页无跨域设置
      favicon: "" #网页图标url,通过 301 跳转获取,暂不支持本地图片,请使用在线资源
      beianMiit: "" #备案号
```
## build for source

```shell
git clone https://github.com/xinjiajuan/WebDav-WebClient_go.git
cd WebDav-WebClient_go
go mod tidy
go build
```