package config

import (
	"net/url"
)

type config struct {
	TargetURL url.URL `env:"TARGET_URL,required,notEmpty"`
	AuthToken string  `env:"AUTH_TOKEN,required,notEmpty"`
}
