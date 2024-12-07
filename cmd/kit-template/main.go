// Package main contains the kit-template service binary.
package main

import (
	"context"
	"fmt"
	"net"

	"github.com/bool64/ctxd"
	logicalservices "github.com/dohernandez/go-grpc-service"
	sapp "github.com/dohernandez/go-grpc-service/app"
	sconfig "github.com/dohernandez/go-grpc-service/config"
	"github.com/dohernandez/go-grpc-service/must"
	"github.com/dohernandez/kit-template/internal/platform/app"
	"github.com/dohernandez/kit-template/internal/platform/config"
	"github.com/dohernandez/servers"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// load configurations
	var cfg config.Config

	err := sconfig.LoadConfig(&cfg)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load configurations"))

	metricsListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppMetricsPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init Metrics service listener"))

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppGRPCPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init GRPC service listener"))

	optReflection := func(any) {}

	if cfg.IsDev() {
		optReflection = servers.WithReflection()
	}

	restTListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppRESTPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service listener"))

	// initialize locator
	deps, err := app.NewServiceLocator(
		&cfg,
		sapp.WithGRPC(
			optReflection,
			servers.WithListener(grpcListener, true),
		),
		sapp.WithGRPCRest(
			servers.WithListener(restTListener, true),
		),
		sapp.WithMetrics(
			servers.WithListener(metricsListener, true),
		),
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init locator"))

	err = logicalservices.RunServices(ctx, deps.Locator)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to start the services"))
}
