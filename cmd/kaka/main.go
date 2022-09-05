package main

import (
	"flag"
	"kaka/internal/log"
	"os"
)

var (
	Name       = "kaka"
	Version    = "v0.0.1"
	flagConfig string
	id, _      = os.Hostname()
)

func init() {
	flag.StringVar(&flagConfig, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller)
	l := log.NewHelper(logger)
	l.Infof("can not create app: %v", "err: can not open database")
}
