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
