package main

import (
	"flag"
	"fmt"
	"github.com/ChinasMr/kaka/internal/conf"
	"github.com/ChinasMr/kaka/internal/config"
	"github.com/ChinasMr/kaka/internal/config/file"
	"github.com/ChinasMr/kaka/internal/log"
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
	c := config.New(
		config.WithSource(
			file.NewSource(flagConfig),
		),
	)
	err := c.Load()
	if err != nil {
		panic(err)
	}
	var bc conf.Bootstrap
	err = c.Scan(&bc)
	if err != nil {
		panic(err)
	}

	err = c.Watch("server.rtsp.port", func(s string, value config.Value) {
		i, _ := value.Int()

		fmt.Printf("new rtsp port is %d", i)
	})
	if err != nil {
		panic(err)
	}
	ch := make(chan bool)
	ch <- true
}
