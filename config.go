package main

import r53types "github.com/aws/aws-sdk-go-v2/service/route53/types"

type hostedZoneInfo struct {
	Id   string
	Name string
}

type recordSetInfo struct {
	Name          string
	Type          r53types.RRType
	SetIdentifier string
	Weight        int64
}

type Config struct {
	AppName   string `env:"APP_NAME,default=dns-switchover"`
	Owner     string `env:"OWNER,default=mesh@dazn.com"`
	AwsRegion string `env:"AWS_REGION,default=eu-central-1"`
}
