package main

import (
	"context"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/kit-template/internal/platform/app"
	"github.com/dohernandez/kit-template/internal/platform/config"
	grpcServer "github.com/dohernandez/kit-template/pkg/grpc/server"
	"github.com/dohernandez/kit-template/pkg/must"
	"github.com/dohernandez/kit-template/pkg/servicing"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// load configurations
	cfg, err := config.GetConfig()
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load configurations"))

	srvMetrics, err := app.NewMetricsService(ctx, cfg)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init Metrics service"))

	// initialize locator
	deps, err := app.NewServiceLocator(cfg, func(l *app.Locator) {
		l.GRPCUnitaryInterceptors = append(l.GRPCUnitaryInterceptors,
			// adding metrics
			srvMetrics.ServerMetrics().UnaryServerInterceptor(),
		)
	})
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init locator"))

	srvGRPC, srv, err := app.NewGRPCService(ctx, cfg, deps, grpcServer.WithMetrics(srvMetrics.ServerMetrics()))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init GRPC service"))

	srvREST, err := app.NewRESTService(ctx, cfg, deps, srv)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service"))

	services := servicing.WithGracefulSutDown(
		func(ctx context.Context) {
			app.GracefulDBShutdown(ctx, deps)
		},
	)

	err = services.Start(
		ctx,
		func(ctx context.Context, msg string) {
			deps.CtxdLogger().Important(ctx, msg)
		},
		srvMetrics,
		srvGRPC,
		srvREST,
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to start the services"))
}
