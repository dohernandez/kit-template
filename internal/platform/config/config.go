package config

import (
	sapp "github.com/dohernandez/go-grpc-service/app"
)

// Config represents config with variables needed for an app.
type Config struct {
	*sapp.Config

	// Add your custom config variables here.
}
