package Log

import (
	"github.com/sirupsen/logrus"
)

var AppLog = logrus.New()
var logFormatter = logrus.TextFormatter{
	ForceColors:            true,
	FullTimestamp:          true,
	DisableLevelTruncation: true,
	PadLevelText:           true,
}

func InitLog() {
	//AppLog.SetFormatter(&logFormatter)
	AppLog.SetFormatter(&logFormatter)
}
func SetReportCaller(bo bool) {
	AppLog.SetReportCaller(bo)
}
