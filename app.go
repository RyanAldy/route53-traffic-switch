package main

import (
	"context"
	"log"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

type App struct {
	config                  Config
	route53Client           IRoute53
	oldClusterSuffix        *string
	newClusterSuffix        *string
	trafficSwitchPercentage *int64
	region                  *string
	environment             *string
}

func NewApp(config Config, oldClusterSuffix *string, newClusterSuffix *string, trafficSwitchPercentage *int64, region *string, environment *string) (*App, error) {
	app := &App{
		config:                  config,
		oldClusterSuffix:        oldClusterSuffix,
		newClusterSuffix:        newClusterSuffix,
		trafficSwitchPercentage: trafficSwitchPercentage,
		region:                  region,
		environment:             environment,
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(config.AwsRegion))
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	app.route53Client = route53.NewFromConfig(cfg)

	return app, nil
}
