package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type mockRoute53Client struct {
	route53iface.Route53API
}

func TestHostedZoneResult(t *testing.T) {

	main()
}
