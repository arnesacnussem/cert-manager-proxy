package main

import (
	"acmeproxy/proxy"
	"github.com/sirupsen/logrus"
	"runtime"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: false,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return f.Function, ""
		},
	})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
}

func main() {
	server := proxy.NewServer()
	server.Serve()
}
