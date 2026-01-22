package config

import (
	"errors"
	"net/url"
	"os"
	"reflect"

	"main/internal/logger"

	"github.com/caarlos0/env/v11"
)

var Global = config{}

func init() {
	errs := []error{}

	if err := env.ParseWithOptions(&Global, env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(url.URL{}): func(envVar string) (any, error) {
				if u, err := url.ParseRequestURI(envVar); err != nil {
					return nil, err
				} else {
					return *u, nil
				}
			},
		},
	}); err != nil {
		if er, ok := err.(env.AggregateError); ok {
			errs = append(errs, er.Errors...)
		} else {
			errs = append(errs, err)
		}
	}

	if len(Global.AuthToken) < 16 {
		errs = append(errs, errors.New("AUTH_TOKEN must be at least 16 characters"))
	}

	if len(errs) > 0 {
		logger.Stderr.Error("error parsing environment variables", logger.ErrorsAttr(errs...))

		os.Exit(1)
	}
}
