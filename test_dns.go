package main

import (
	"context"
	"testing"
	"time"
	// "github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go/aws"

	r53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// type mockRoute53Client struct {
// 	route53iface.Route53API
// }

type mockRoute53Client struct{}

func (r mockRoute53Client) ChangeResourceRecordSets(ctx context.Context, input *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
	timeNow := time.Now()
	return &route53.ChangeResourceRecordSetsOutput{
		ChangeInfo: &r53types.ChangeInfo{
			Id:          aws.String("id123"),
			Status:      r53types.ChangeStatusInsync,
			SubmittedAt: &timeNow,
		},
	}, nil
}

func TestHostedZoneResult(t *testing.T) {

	main()
}
