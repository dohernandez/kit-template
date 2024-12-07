package app

import (
	sapp "github.com/dohernandez/go-grpc-service/app"
	"github.com/dohernandez/kit-template/internal/platform/config"
	"github.com/dohernandez/kit-template/internal/platform/service"
	"github.com/dohernandez/kit-template/resources/swagger"
	"github.com/dohernandez/servers"
)

// Locator defines application resources.
type Locator struct {
	*sapp.Locator

	cfg *config.Config

	KitTemplateService *service.KitTemplateService

	// use cases
}

// NewServiceLocator creates application locator.
func NewServiceLocator(cfg *config.Config, opts ...sapp.Option) (*Locator, error) {
	// Init postgres database
	opts = append(opts, sapp.WithPostgresDBx())

	upl, err := sapp.NewServiceLocator(cfg.Config, opts...)
	if err != nil {
		return nil, err
	}

	l := &Locator{
		Locator: upl,
		cfg:     cfg,
	}

	l.setupStorage()

	// setting up use cases dependencies
	l.setupUsecaseDependencies()

	l.KitTemplateService = service.NewKitTemplateService(l)

	err = l.setupServices()
	if err != nil {
		return nil, err
	}

	return l, nil
}

// setupStorage sets up storage dependencies (platform).
func (l *Locator) setupStorage() {}

// setupUsecaseDependencies sets up use case dependencies (domain).
func (l *Locator) setupUsecaseDependencies() {}

func (l *Locator) setupServices() error {
	l.InitGRPCService(
		servers.WithRegisterService(l.KitTemplateService),
	)

	err := l.InitGRPCRestService(
		servers.WithRegisterServiceHandler(l.KitTemplateService),
		servers.WithDocEndpoint(l.cfg.ServiceName,
			"/docs/",
			"/docs/service.swagger.json",
			swagger.SwgJSON),
		servers.WithVersionEndpoint(),
	)
	if err != nil {
		return err
	}

	l.InitMetricsService(servers.WithGRPCServer(l.GRPCService))

	return nil
}
