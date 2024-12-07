package service

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// RegisterService registers the service implementation to grpc service.
func (s *KitTemplateService) RegisterService(_ grpc.ServiceRegistrar) {
	// register grpc service
}

// RegisterServiceHandler registers the service implementation to mux.
func (s *KitTemplateService) RegisterServiceHandler(_ *runtime.ServeMux) error {
	// Uncomment this line once the grpc files were generated into the proto package.
	// register rest service
	// return api.RegisterKitTemplateServiceHandlerFromEndpoint(context.Background(), mux, s.deps.GRPCAddr(), []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	return nil
}
