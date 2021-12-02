package service

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// RegisterHandlerService registers the service implementation to mux.
func (s *KitTemplateRESTService) RegisterHandlerService(mux *runtime.ServeMux) error {
	// register rest service
	return nil
}
