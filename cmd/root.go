package cmd

import (
	"github.com/KeisukeYamashita/slackduty/client"
	"github.com/KeisukeYamashita/slackduty/config"
	"github.com/KeisukeYamashita/slackduty/log"
	"go.uber.org/zap"
)

// Execute will run the slackduty job
func Execute() error {
	logger, err := log.New("INFO")
	if err != nil {
		return err
	}

	return execute(logger)
}

func execute(logger *zap.Logger) error {
	path := config.GetConfigPath()
	cfg, err := config.Load(path)
	if err != nil {
		logger.Error("failed to load config", zap.Error(err))
		return err
	}

	pdAPIKey, slackAPIKey, err := config.GetAPIKeys()
	if err != nil {
		logger.Error("failed to load API key", zap.Error(err))
		return err
	}

	opts := []client.ClientOption{}

	if config.IsExternalTrigger() {
		opts = append(opts, client.WithExternalTrigger())
	}

	client := client.New(cfg, pdAPIKey, slackAPIKey, logger, opts...)
	return client.Run()
}
