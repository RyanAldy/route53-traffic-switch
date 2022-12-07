package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

type App struct {
	// config        *config.Config
	route53Client IRoute53
}

func New() (*App, error) {
	app := &App{}

	// To be parameterised also - region
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-central-1"))
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	app.route53Client = route53.NewFromConfig(cfg)

	return app, nil
}
