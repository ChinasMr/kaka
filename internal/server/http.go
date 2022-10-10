package server

import (
	v1 "github.com/ChinasMr/kaka/api/kaka/v1"
	"github.com/ChinasMr/kaka/internal/conf"
	"github.com/ChinasMr/kaka/internal/service"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// protoc plugins is small complicate. directly use kratos temporary.
// need time to read and understand the source code of protoc-gen-go-http plugins.
// still too young. struggle for !!!

func NewHttpServer(c *conf.Server, kaka *service.KakaService) *http.Server {
	var opts []http.ServerOption
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterKakaHTTPServer(srv, kaka)
	return srv
}
