package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/kit-template/internal/platform/app"
	"github.com/dohernandez/kit-template/internal/platform/config"
	"github.com/dohernandez/kit-template/internal/platform/service"
	grpcMetrics "github.com/dohernandez/kit-template/pkg/grpc/metrics"
	grpcRest "github.com/dohernandez/kit-template/pkg/grpc/rest"
	grpcServer "github.com/dohernandez/kit-template/pkg/grpc/server"
	"github.com/dohernandez/kit-template/pkg/must"
	"github.com/dohernandez/kit-template/resources/swagger"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcZapLogger "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcCtxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpcOpentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	mux "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v3 "github.com/swaggest/swgui/v3"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// load configurations
	cfg, err := config.GetConfig()
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load configurations"))

	// initialize locator
	l, err := app.NewServiceLocator(cfg)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init locator"))

	// grpc service
	s := &service.KitTemplateService{}

	// register
	grpcServiceRegister := service.NewGRPCServiceRegister(s)
	restServiceRegister := service.NewRESTServiceRegister(s)

	errCh := make(chan error, 1)

	// Enabling graceful shutdown
	toShutdown, shutdownCh, shutdownDoneCh := app.GracefulShutdown(ctx, l, errCh)

	// starting metrics service
	metricsShutdownDoneCh := make(chan struct{})
	srvMetrics := startMetricsService(ctx, cfg, l, shutdownCh, metricsShutdownDoneCh, errCh)
	toShutdown["metrics"] = metricsShutdownDoneCh

	// enabling interceptor for grpc and rest
	interceptors := enablingGRPCInterceptors(l, srvMetrics)

	// starting grpc service
	grpcShutdownDoneCh := make(chan struct{})
	startGRPCService(ctx, cfg, l, grpcServiceRegister, interceptors, srvMetrics, shutdownCh, grpcShutdownDoneCh, errCh)
	toShutdown["grpc"] = grpcShutdownDoneCh

	// starting rest service
	restShutdownDoneCh := make(chan struct{})
	startRESTService(ctx, cfg, l, restServiceRegister, interceptors, shutdownCh, restShutdownDoneCh, errCh)
	toShutdown["rest"] = restShutdownDoneCh

	for {
		select {
		case err := <-errCh:
			must.NotFail(err)

		case <-shutdownDoneCh:
			return
		}
	}
}

func enablingGRPCInterceptors(l *app.Locator, srvMetrics *grpcMetrics.Server) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		// recovering from panic
		grpcRecovery.UnaryServerInterceptor(),
		// adding tracing
		grpcOpentracing.UnaryServerInterceptor(),
		// adding metrics
		srvMetrics.ServerMetrics().UnaryServerInterceptor(),
		// adding logger
		grpcCtxtags.UnaryServerInterceptor(grpcCtxtags.WithFieldExtractor(grpcCtxtags.CodeGenRequestFieldExtractor)),
		grpcZapLogger.UnaryServerInterceptor(l.ZapLogger()),
	}
}

func startGRPCService(
	ctx context.Context,
	cfg *config.Config,
	locator *app.Locator,
	grpcServiceRegister *service.KitTemplateGRPCServiceRegister,
	interceptors []grpc.UnaryServerInterceptor,
	metricsServer *grpcMetrics.Server,
	shutdownCh chan struct{},
	shutdownDoneCh chan struct{},
	errCh chan error,
) {
	GRPCListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppGRPCPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed starting GRPC listener"))

	grpcZapLogger.ReplaceGrpcLoggerV2(locator.ZapLogger())

	opts := []grpcServer.Option{
		grpcServer.WithListener(GRPCListener, true),
		// registering point service using the point service registerer
		grpcServer.WithService(grpcServiceRegister),
		grpcServer.WithMetrics(metricsServer.ServerMetrics()),
		grpcServer.ChainUnaryInterceptor(interceptors...),
	}

	if cfg.IsDev() {
		opts = append(opts, grpcServer.WithReflective())
	}

	srv := grpcServer.NewServer(opts...)

	go func() {
		locator.CtxdLogger().Important(context.Background(), fmt.Sprintf("start GRPC server at addr %s", GRPCListener.Addr().String()))

		errCh <- srv.WithShutdownSignal(shutdownCh, shutdownDoneCh).Start()
	}()
}

func startRESTService(
	ctx context.Context,
	cfg *config.Config,
	locator *app.Locator,
	restServiceRegister *service.KitTemplateRESTServiceRegister,
	interceptors []grpc.UnaryServerInterceptor,
	shutdownCh chan struct{},
	shutdownDoneCh chan struct{},
	errCh chan error,
) {
	RESTListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppRESTPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed starting REST listener"))

	restServiceRegister.WithUnaryServerInterceptor(
		grpcMiddleware.ChainUnaryServer(interceptors...),
	)

	opts := []grpcRest.Option{
		grpcRest.WithListener(RESTListener, true),
		// use to registering point service using the point service registerer
		grpcRest.WithService(restServiceRegister),

		// handler root path
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				w.Header().Set("content-type", "text/html")

				_, err := w.Write([]byte("Welcome to " + locator.Config.ServiceName +
					`. Please read API <a href="docs">documentation</a>.`))
				if err != nil {
					locator.CtxdLogger().Error(r.Context(), "failed to write response",
						"error", err)
				}
			})
		}),
	}

	swaggerOptions := swaggerHandlersOptions(ctx, locator)
	opts = append(opts, swaggerOptions...)

	rest, err := grpcRest.NewServer(opts...)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service"))

	go func() {
		locator.CtxdLogger().Important(context.Background(), fmt.Sprintf("start REST server at addr %s", RESTListener.Addr().String()))

		errCh <- rest.WithShutdownSignal(shutdownCh, shutdownDoneCh).Start()
	}()
}

func swaggerHandlersOptions(ctx context.Context, locator *app.Locator) []grpcRest.Option {
	swh := v3.NewHandler(locator.Config.ServiceName, "/docs/service.swagger.json", "/docs/")

	return []grpcRest.Option{
		// handler docs paths
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/docs/service.swagger.json", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				w.Header().Set("Content-Type", "application/json")

				_, err := w.Write(swagger.SwgJSON)
				must.NotFail(ctxd.WrapError(ctx, err, "failed to load /docs/service.swagger.json file"))
			})
		}),
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/docs", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			})
		}),
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/docs/swagger-ui-bundle.js", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			})
		}),
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/docs/swagger-ui-standalone-preset.js", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			})
		}),
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/docs/swagger-ui.css", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			})
		}),
	}
}

func startMetricsService(
	ctx context.Context,
	cfg *config.Config,
	locator *app.Locator,
	shutdownCh chan struct{},
	shutdownDoneCh chan struct{},
	errCh chan error,
) *grpcMetrics.Server {
	MetricsListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppMetricsPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed starting REST listener"))

	opts := []grpcMetrics.Option{
		grpcMetrics.WithListener(MetricsListener, true),
	}

	metrics := grpcMetrics.NewServer(opts...)

	go func() {
		locator.CtxdLogger().Important(context.Background(), fmt.Sprintf("start Metrics server at addr %s", MetricsListener.Addr().String()))

		errCh <- metrics.WithShutdownSignal(shutdownCh, shutdownDoneCh).Start()
	}()

	return metrics
}
