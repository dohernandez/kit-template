package app

import (
	"context"
	"fmt"
	"net"

	"github.com/dohernandez/kit-template/internal/platform/config"
	"github.com/dohernandez/kit-template/internal/platform/service"
	grpcServer "github.com/dohernandez/kit-template/pkg/grpc/server"
	grpcZapLogger "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
)

// NewGRPCService creates an instance of grpc service, with all the instrumentation.
func NewGRPCService(
	_ context.Context,
	cfg *config.Config,
	locator *Locator,
	opts ...grpcServer.Option,
) (*grpcServer.Server, *service.KitTemplateService, error) {
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppGRPCPort))
	if err != nil {
		return nil, nil, err
	}

	srv := service.NewKitTemplateService()

	grpcZapLogger.ReplaceGrpcLoggerV2(locator.ZapLogger())

	opts = append(opts,
		grpcServer.WithListener(grpcListener, true),
		// registering point service using the point service registerer
		grpcServer.WithService(srv),
		grpcServer.ChainUnaryInterceptor(locator.GRPCUnitaryInterceptors...),
	)

	// Enabling reflection in dev and testing env.
	if cfg.IsDev() || cfg.IsTest() {
		opts = append(opts, grpcServer.WithReflective())
	}

	return grpcServer.NewServer(opts...), srv, nil
}
