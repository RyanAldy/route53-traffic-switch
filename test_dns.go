package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go/aws"

	// "github.com/stretchr/testify/assert"

	r53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

type mockRoute53Client struct {
}

// type mockRoute53Client struct{}

// ctx := context.TODO()

func (r *mockRoute53Client) ChangeResourceRecordSets(ctx context.Context, input *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {

	timeNow := time.Now()
	return &route53.ChangeResourceRecordSetsOutput{
		ChangeInfo: &r53types.ChangeInfo{
			Id:          aws.String("id123"),
			Status:      r53types.ChangeStatusInsync,
			SubmittedAt: &timeNow,
		},
	}, nil
}

// func TestModifyRecords(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		action        string
// 		errorExpected bool
// 	}{
// 		{
// 			name:          "Modify Route53 records - Upsert",
// 			action:        "upsert",
// 			errorExpected: false,
// 		},
// 		{
// 			name:          "Modify Route53 records - Delete",
// 			action:        "delete",
// 			errorExpected: false,
// 		},
// 		{
// 			name:          "Modify Route53 records - invalid action",
// 			action:        "invalid",
// 			errorExpected: true,
// 		},
// 	}
// 	// config, _ := config.New()
// 	// app := &App{
// 	// 	route53Client: new(mockRoute53Client),
// 	// 	config:        config,
// 	// }

// 	mockR53Client := &mockRoute53Client{}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := switchTraffic(mockR53Client, "test-service")
// 			if tt.errorExpected {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }
