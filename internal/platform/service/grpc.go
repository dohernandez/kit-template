package service

import (
	"github.com/bool64/ctxd"
)

// KitTemplateServiceDeps holds the dependencies for the KitTemplateService.
type KitTemplateServiceDeps interface {
	Logger() ctxd.Logger
	GRPCAddr() string
}

// KitTemplateService is the gRPC service.
type KitTemplateService struct {
	// Uncomment this line once the grpc files were generated into the proto package.
	// UnimplementedKitTemplateServiceServer must be embedded to have forward compatible implementations.
	// api.UnimplementedKitTemplateServiceServer

	deps KitTemplateServiceDeps
}

// NewKitTemplateService creates a new KitTemplateService.
func NewKitTemplateService(deps KitTemplateServiceDeps) *KitTemplateService {
	return &KitTemplateService{
		deps: deps,
	}
}

/*
// PostFuncName ... .
func (s *KitTemplateService) PostFuncName(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}
*/
