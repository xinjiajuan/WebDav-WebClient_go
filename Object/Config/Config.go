package Config

import (
	"WebDav-ClientWeb/Object/Config/Log"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func ReadConfig(path string) Yaml {
	var conf Yaml
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		Log.SetReportCaller(true)
		Log.AppLog.Fatalln(err.Error())
		Log.SetReportCaller(false)
	}
	Log.AppLog.Infoln("Config Read is OK!")
	if err = yaml.Unmarshal(yamlFile, &conf); err != nil {
		Log.SetReportCaller(true)
		Log.AppLog.Fatalln(err.Error())
		Log.SetReportCaller(false)
	}
	Log.AppLog.Infoln("Config Unmarshal is OK!")
	return conf
}
