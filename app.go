package main

import (
	"context"
	"log"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

type App struct {
	config        Config
	route53Client IRoute53
}

func NewApp(config Config) (*App, error) {
	app := &App{
		config: config,
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(config.AwsRegion))
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	app.route53Client = route53.NewFromConfig(cfg)

	return app, nil
}
