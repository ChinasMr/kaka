package main

import (
	"flag"
	"github.com/ChinasMr/kaka/internal/conf"
	"github.com/ChinasMr/kaka/pkg/application"
	"github.com/ChinasMr/kaka/pkg/config"
	"github.com/ChinasMr/kaka/pkg/config/file"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/ChinasMr/kaka/pkg/transport/grpc"
	"github.com/ChinasMr/kaka/pkg/transport/rtsp"
	"github.com/go-kratos/kratos/v2/transport/http"
	"os"
)

var (
	Name       = "kaka"
	Version    = "alpha0.3"
	flagConfig string
	id, _      = os.Hostname()
)

func init() {
	flag.StringVar(&flagConfig, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, ht *http.Server, rtsp *rtsp.Server) *application.App {
	return application.New(
		application.ID(id),
		application.Name(Name),
		application.Version(Version),
		application.Metadata(map[string]string{}),
		application.Logger(logger),
		application.Server(
			gs,
			rtsp,
			ht,
		),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller)

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

	app, cleanup, err := wireApp(bc.Server, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	err = app.Run()
	if err != nil {
		panic(err)
	}
}
