package Config

import (
	"time"
)

type Yaml struct {
	ServerList []Server `yaml:"server"`
}
type Server struct {
	Name       string `yaml:"Webdavname"`
	ListenPort int    `yaml:"listenPort"`
	Enable     bool   `yaml:"enable"`
	Host       string `yaml:"webdavhost"`
	UserName   string `yaml:"user"`
	PassWD     string `yaml:"passwd"`
	Options    struct {
		UseTLS struct {
			Enable   bool   `yaml:"enable"`
			CertFile string `yaml:"certFile"`
			CertKey  string `yaml:"certKey"`
		} `yaml:"useTLS"`
		AccessControlAllowOrigin string `yaml:"access-control-allow-origin"`
		Favicon                  string `yaml:"favicon"`
		BeianMiit                string `yaml:"beianMiit"`
	} `yaml:"weboptions"`
}

type ObjectInfo struct {
	Key          string
	Size         int64
	IsDir        bool
	LsatModified time.Time
}

var (
	Version string = "1.0.1"
	AppName string = "WebDav-WebClient"
	Usage   string = "基于WebDav文件下载服务器"
)
