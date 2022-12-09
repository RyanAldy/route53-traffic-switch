package main

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Env          string `env:"ENV" envDefault:"dev"`
	AppName      string `env:"APP_NAME,default=dns-switchover"`
	Owner        string `env:"OWNER,default=mesh@dazn.com"`
	AwsRegion    string `env:"AWS_REGION,default=eu-central-1"`
	HostedZoneID string `env:"HOSTED_ZONE_ID,required"`
}

func New() (*Config, error) {
	var c Config
	if err := envconfig.Process(context.Background(), &c); err != nil {
		return nil, errors.Wrapf(err, "failed to process config")
	}

	return &c, nil
}
