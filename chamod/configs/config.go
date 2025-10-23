package configs

type Config struct {
	Port           int            `mapstructure:"port"`
	NewRelicConfig NewRelicConfig `mapstructure:"newRelicConfig"`
}

type NewRelicConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	AppName    string `mapstructure:"appName"`
	LicenseKey string `mapstructure:"licenseKey"`
}
