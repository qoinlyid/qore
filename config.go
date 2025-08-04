package qore

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	// App config.
	AppName          string `json:"APP_NAME" mapstructure:"APP_NAME"`
	AppProduction    bool   `json:"APP_PRODUCTION" mapstructure:"APP_PRODUCTION"`
	AppContainerized bool   `json:"APP_CONTAINERIZED" mapstructure:"APP_CONTAINERIZED"`
	ShutdownTimeout  int    `json:"APP_SHUTDOWN_TIMEOUT" mapstructure:"APP_SHUTDOWN_TIMEOUT"`

	// Log config.
	LogLevel      LogLevel `json:"LOG_LEVEL" mapstructure:"LOG_LEVEL"`
	LogJSON       bool     `json:"LOG_JSON" mapstructure:"LOG_JSON"`
	LogShowSource bool     `json:"LOG_SHOW_SOURCE" mapstructure:"LOG_SHOW_SOURCE"`

	// HTTP Server config.
	HTTPPort     int    `json:"HTTP_PORT" mapstructure:"HTTP_PORT"`
	HTTPAutoTLS  bool   `json:"HTTP_AUTO_TLS" mapstructure:"HTTP_AUTO_TLS"`
	HTTPCertPath string `json:"HTTP_CERT_PATH" mapstructure:"HTTP_CERT_PATH"`
	HTTPKeyPath  string `json:"HTTP_KEY_PATH" mapstructure:"HTTP_KEY_PATH"`
}

var defaultConfig = &Config{
	// App.
	AppName:          "qore-app",
	AppContainerized: true,
	ShutdownTimeout:  30,

	// Log.
	LogLevel:      LOG_DEBUG,
	LogShowSource: true,

	// HTTP.
	HTTPPort: 3100,
}

func loadConfig() *Config {
	var e error
	config := defaultConfig

	// Get used config from OS env.
	configSource := os.Getenv(CONFIG_USED_KEY)
	if ValidationIsEmpty(configSource) {
		configSource = "OS"
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	switch strings.ToUpper(configSource) {
	case "OS":
		if err := viper.Unmarshal(&config); err != nil {
			e = errors.Join(fmt.Errorf("failed to parse OS env value to config: %w", err))
		}
	default:
		ext := strings.ToLower(filepath.Ext(configSource))
		switch ext {
		case ".env":
			if err := godotenv.Load(configSource); err != nil {
				e = errors.Join(fmt.Errorf("failed to load .env file %s: %w", configSource, err))
			}
			if err := viper.Unmarshal(&config); err != nil {
				e = errors.Join(fmt.Errorf("failed to parse .env value to config: %w", err))
			}
		case ".json", ".yml", ".yaml", ".toml":
			viper.SetConfigFile(configSource)
			if err := viper.ReadInConfig(); err != nil {
				e = errors.Join(fmt.Errorf("failed to read config file %s: %w", configSource, err))
			} else {
				if err := viper.Unmarshal(&config); err != nil {
					e = errors.Join(fmt.Errorf("failed to parse config file %s value to config: %w", configSource, err))
				}
			}
		}
	}

	if e != nil {
		log.Printf("qore config - failed to load config: %s\n", e.Error())
	}
	return config
}
