package service

import (
	"context"
	"errors"
)

var errNotImplemented = errors.New("not implemented")

// KitTemplateService ... .
type KitTemplateService struct{}

// PostFuncName ... .
func (s *KitTemplateService) PostFuncName(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, errNotImplemented
}
