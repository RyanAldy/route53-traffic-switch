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

func (r mockRoute53Client) ChangeResourceRecordSets(ctx context.Context, input *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {

	timeNow := time.Now()
	return &route53.ChangeResourceRecordSetsOutput{
		ChangeInfo: &r53types.ChangeInfo{
			Id:          aws.String("test123"),
			Status:      r53types.ChangeStatusInsync,
			SubmittedAt: &timeNow,
		},
	}, nil
}

func (r mockRoute53Client) ListResourceRecordSets(ctx context.Context, params *route53.ListResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
	return &route53.ListResourceRecordSetsOutput{}, nil
}

func TestUpdateRecords(t *testing.T) {
	var recordInfo = []recordSetInfo{
		{"euc1.internal-dev.dazn-gateway.com.", r53types.RRType("A"), "mesh-657", 0},
		{"euc1.internal-dev.dazn-gateway.com.", r53types.RRType("A"), "mesh-696", 0},
		{"euc1.internal-dev.dazn-gateway.com.", r53types.RRType("A"), "mesh-697", 0},
		{"euc1.internal-dev.dazn-gateway.com.", r53types.RRType("AAAA"), "mesh-657", 0},
		{"euc1.internal-dev.dazn-gateway.com.", r53types.RRType("AAAA"), "mesh-696", 0},
		{"{euc1.internal-dev.dazn-gateway.com.", r53types.RRType("AAAA"), "mesh-697", 0},
	}
	tests := []struct {
		name                  string
		errorExpected         bool
		old                   string
		new                   string
		records               []recordSetInfo
		weight                int64
		inputWeightPercentage int64
		recordType            string
	}{
		{
			name:                  "Modify Route53 records - Upsert A type",
			errorExpected:         false,
			old:                   "mesh-657",
			new:                   "mesh-697",
			records:               recordInfo,
			weight:                255,
			inputWeightPercentage: 50,
			recordType:            "A",
		},
		{
			name:                  "Modify Route53 records - Upsert AAAA type",
			errorExpected:         false,
			old:                   "mesh-657",
			new:                   "mesh-697",
			records:               recordInfo,
			weight:                0,
			inputWeightPercentage: 100,
			recordType:            "AAAA",
		},
		{
			name:                  "Modify Route53 records - test incorrect cluster name input",
			errorExpected:         true,
			old:                   "mesh-xxx",
			new:                   "mesh-697",
			records:               recordInfo,
			weight:                255,
			inputWeightPercentage: 50,
			recordType:            "AAAA",
		},
		// Need to think of a way of testing this
		// {
		// 	name:                  "Modify Route53 records - test incorrect weight input",
		// 	errorExpected:         true,
		// 	old:                   "mesh-657",
		// 	new:                   "mesh-697",
		// 	records:               recordInfo,
		// 	weight:                255,
		// 	inputWeightPercentage: 110,
		// 	recordType:            "A",
		// },
	}

	config, _ := New()
	app, _ := NewApp(*config, nil, nil, nil, nil, nil)
	mockR53Client := &mockRoute53Client{}
	app.route53Client = mockR53Client

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.switchTraffic(tt.records, tt.old, tt.new, tt.weight, tt.inputWeightPercentage, tt.recordType)
			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
