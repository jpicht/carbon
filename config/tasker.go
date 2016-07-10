package config

import (
	"encoding/json"

	"golang.org/x/net/context"
)

type Tasker interface {
	json.Marshaler
	json.Unmarshaler

	Run(context.Context, *BuildConfig)
}

type BuildConfig struct {
	TempDirectory   string
	TargetDirectory string
}
