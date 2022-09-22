//go:build wireinject
// +build wireinject

package main

import (
	"github.com/ChinasMr/kaka/internal/server"
	"github.com/ChinasMr/kaka/internal/service"
	"github.com/ChinasMr/kaka/pkg/log"
	"github.com/google/wire"
)

import (
	"github.com/ChinasMr/kaka/internal/conf"
	"github.com/ChinasMr/kaka/pkg/app"
)

func wireApp(*conf.Server, log.Logger) (*app.App, func(), error) {
	panic(wire.Build(newApp, server.ProviderSet, service.ProviderSet))
}
