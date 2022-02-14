package service

import (
	"errors"
)

var errNotImplemented = errors.New("not implemented")

// KitTemplateService ... .
type KitTemplateService struct {
	// Uncomment this line once the grpc files were generated into the proto package
	// UnimplementedKitTemplateServiceServer must be embedded to have forward compatible implementations.
	// api.UnimplementedKitTemplateServiceServer
}

// NewKitTemplateService ...
func NewKitTemplateService() *KitTemplateService {
	return &KitTemplateService{}
}

/*
// PostFuncName ... .
func (s *KitTemplateService) PostFuncName(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, errNotImplemented
}
*/
