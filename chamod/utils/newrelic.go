package utils

import (
	"chamod/configs"
	"fmt"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func SetupNewRelic(config *configs.Config) *newrelic.Application {
	Logger.Info("", "Setting up newrelic client")
	if config.NewRelicConfig.Enabled {
		newrelicApp, err := newrelic.NewApplication(
			newrelic.ConfigAppName(config.NewRelicConfig.AppName),
			newrelic.ConfigLicense(config.NewRelicConfig.LicenseKey),
			newrelic.ConfigEnabled(true),
			newrelic.ConfigDistributedTracerEnabled(true),
		)
		if err != nil {
			Logger.Error("", "Error creating newrelic app: %v", err)
			return nil
		}

		Logger.Info("", "New Relic client enabled with agent version: %s, App name: %s", newrelic.Version, config.NewRelicConfig.AppName)

		return newrelicApp
	} else {
		fmt.Printf("New Relic client is disabled")
		return nil
	}
}
