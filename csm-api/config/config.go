package config

import (
	"csm-api/utils"
	"github.com/caarlos0/env"
)

type Config struct {
	Env            string `env:"ENV" envDefault:"local"`
	Role           string `env:"ROLE" envDefault:"web"`
	Port           int    `env:"PORT" envDefault:"8082"`
	Domain         string `env:"DOMAIN" envDefault:"localhost"`
	UploadPath     string `env:"UPLOAD_PATH" envDefault:"uploads"`
	LogPath        string `env:"LOG_PATH" envDefault:"logs"`
	ErrLogPath     string `env:"ERR_LOG_PATH" envDefault:"logs/error"`
	ExcelPath      string `env:"EXCEL_PATH" envDefault:"resources/excel"`
	ConsoleLogPath string `env:"CONSOLE_LOG_PATH" envDefault:"logs/console"`
	SecretKey      string `env:"SECRET_KEY" envDefault:"regno_secret_key"`
}

// caarlos0/env 패키지를 사용하여 struct의 envDefault값을 환경변수로 넘겨준다.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, utils.CustomErrorf(err)
	}
	return cfg, nil
}
