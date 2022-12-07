package main

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"

	r53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

type mockRoute53Client struct{}

// ctx := context.TODO()

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

func (r mockRoute53Client) ListHostedZonesByName(ctx context.Context, params *route53.ListHostedZonesByNameInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
	return &route53.ListHostedZonesByNameOutput{}, nil
}

func (r mockRoute53Client) ListResourceRecordSets(ctx context.Context, params *route53.ListResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
	return &route53.ListResourceRecordSetsOutput{}, nil
}

var recordInfo = []recordSetInfo{
	recordSetInfo{"euc1.internal-dev.dazn-gateway.com.", r53types.RRType("A"), "mesh-657", 0},
	recordSetInfo{"euc1.internal-dev.dazn-gateway.com.", r53types.RRType("A"), "mesh-696", 0},
	recordSetInfo{"euc1.internal-dev.dazn-gateway.com.", r53types.RRType("A"), "mesh-697", 0},
	recordSetInfo{"euc1.internal-dev.dazn-gateway.com.", r53types.RRType("AAAA"), "mesh-657", 0},
	recordSetInfo{"euc1.internal-dev.dazn-gateway.com.", r53types.RRType("AAAA"), "mesh-696", 0},
	recordSetInfo{"{euc1.internal-dev.dazn-gateway.com.", r53types.RRType("AAAA"), "mesh-697", 0},
}

func TestUpdateRecords(t *testing.T) {
	tests := []struct {
		name             string
		errorExpected    bool
		old              string
		new              string
		records          []recordSetInfo
		inputWeight      int64
		weightPercentage int64
		recordType       string
	}{
		{
			name:             "Modify Route53 records - Upsert A type",
			errorExpected:    false,
			old:              "mesh-657",
			new:              "mesh-697",
			records:          recordInfo,
			inputWeight:      50,
			weightPercentage: 255,
			recordType:       "A",
		},
		{
			name:             "Modify Route53 records - Upsert AAAA type",
			errorExpected:    false,
			old:              "mesh-657",
			new:              "mesh-697",
			records:          recordInfo,
			inputWeight:      50,
			weightPercentage: 255,
			recordType:       "AAAA",
		},
	}

	app, _ := New()
	mockR53Client := &mockRoute53Client{}
	app.route53Client = mockR53Client

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.switchTraffic(tt.records, tt.old, tt.new, tt.inputWeight, tt.weightPercentage, tt.recordType)
			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
