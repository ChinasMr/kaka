package main

import (
	"flag"
	"github.com/ChinasMr/kaka/internal/conf"
	"github.com/ChinasMr/kaka/pkg/app"
	"github.com/ChinasMr/kaka/pkg/config"
	"github.com/ChinasMr/kaka/pkg/config/file"
	"github.com/ChinasMr/kaka/pkg/log"
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

func newApp(logger log.Logger) *app.App {
	return app.New(
		app.ID(id),
		app.Name(Name),
		app.Version(Version),
		app.Metadata(map[string]string{}),
		app.Logger(logger),
		app.Server(),
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
