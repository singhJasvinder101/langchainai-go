package initializers

import (
	"context"

	"github.com/singhJasvinder101/langchainai-go/init/config"
	"github.com/singhJasvinder101/langchainai-go/internal/llm/providers"
	"github.com/singhJasvinder101/langchainai-go/pkg/log"
)

func Init(ctx context.Context, configSrc string) error {
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

	log.WithContext(ctx).Info("initializing provider factory")
	providers.NewProviderFactory(ctx)
	log.WithContext(ctx).Info("provider factory initialization complete")

	return nil
}
