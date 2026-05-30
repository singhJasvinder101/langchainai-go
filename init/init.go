package initializers

import (
	"context"

	"github.com/singhJasvinder101/agentic-go/init/config"
	"github.com/singhJasvinder101/agentic-go/pkg/log"
)

func Init(ctx context.Context, configSrc string) {
	log.Init(log.Options{Level: "info", Format: "json"})

	log.WithContext(ctx).Info("initializing config")
	config.MustInit(configSrc)
	log.WithContext(ctx).Info("config initialization complete")

	log.WithContext(ctx).Info("reconfiguring log")
	log.Reconfigure(log.Options{
		Level:  config.GetString("log.level"),
		Format: config.GetString("log.format"),
	})
	log.WithContext(ctx).Info("log reconfiguration complete")

}
