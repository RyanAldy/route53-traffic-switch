package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/pkg/errors"

	r53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

type recordSetInfo struct {
	Name          string
	Type          r53types.RRType
	SetIdentifier string
	Weight        int64
}

func (a *App) handler() (string, error) {

	ctx := context.TODO()

	trafficWeight, err := convertPerecentageToWeight(*a.trafficSwitchPercentage)
	if err != nil {
		return "", err
	}

	// Will be different for prod - need to add function - intentionally leaving it out for now
	// dnsInput := fmt.Sprintf("%s.dazn-gateway.com", *a.environment)

	// hostedZoneId := "Z080964036XWOHXR8180L"

	hostedZoneId := &a.config.HostedZoneID

	idInput := &route53.ListResourceRecordSetsInput{
		HostedZoneId: hostedZoneId,
	}

	recordSets, err := a.route53Client.ListResourceRecordSets(ctx, idInput)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	recordInfo := []recordSetInfo{}
	for _, record := range recordSets.ResourceRecordSets {
		if *record.Name == fmt.Sprintf("%s.%s.dazn-gateway.com.", *a.region, *a.environment) {
			recordInfo = append(recordInfo, recordSetInfo{
				Name:          *record.Name,
				Type:          record.Type,
				SetIdentifier: *record.SetIdentifier,
				Weight:        *record.Weight})
		}
	}

	startPagiante := recordSets.IsTruncated

	if startPagiante {
		var paginate = true
		paginateIdInput := &route53.ListResourceRecordSetsInput{
			HostedZoneId:          hostedZoneId,
			StartRecordName:       recordSets.NextRecordName,
			StartRecordType:       recordSets.NextRecordType,
			StartRecordIdentifier: recordSets.NextRecordIdentifier,
		}
		newRecordSets, _ := a.route53Client.ListResourceRecordSets(ctx, paginateIdInput)
		for paginate {
			for _, newRecord := range newRecordSets.ResourceRecordSets {
				if *newRecord.Name == fmt.Sprintf("%s.%s.dazn-gateway.com.", *a.region, *a.environment) {
					recordInfo = append(recordInfo, recordSetInfo{
						Name:          *newRecord.Name,
						Type:          newRecord.Type,
						SetIdentifier: *newRecord.SetIdentifier,
						Weight:        *newRecord.Weight})
				}
			}

			paginate = newRecordSets.IsTruncated
			paginateIdInput := &route53.ListResourceRecordSetsInput{
				HostedZoneId:          hostedZoneId,
				StartRecordName:       recordSets.NextRecordName,
				StartRecordType:       recordSets.NextRecordType,
				StartRecordIdentifier: recordSets.NextRecordIdentifier,
			}
			newRecordSets, _ = a.route53Client.ListResourceRecordSets(ctx, paginateIdInput)
		}
	}

	// Can get rid of these and use app ones or keep in for unit testing - will implement this better
	trafficErrA := a.switchTraffic(recordInfo, *a.oldClusterSuffix, *a.newClusterSuffix, trafficWeight, *a.trafficSwitchPercentage, "A")
	if trafficErrA != nil {
		return "", trafficErrA
	}
	trafficErrAAAA := a.switchTraffic(recordInfo, *a.oldClusterSuffix, *a.newClusterSuffix, trafficWeight, *a.trafficSwitchPercentage, "AAAA")
	if trafficErrAAAA != nil {
		return "", trafficErrAAAA
	}

	return "Successfully switched over traffic", nil
}

func (a *App) switchTraffic(records []recordSetInfo, oldClusterSuffix string, newClusterSuffix string, weight int64, weightPercentage int64, recordType string) error {

	if !validateClusterNameInputs(records, oldClusterSuffix) || !validateClusterNameInputs(records, newClusterSuffix) {
		return errors.New("Error updating cluster.  One of the clusters you input possibly does not exist")
	}

	for _, r := range records {
		if r.Type == r53types.RRType(recordType) && strings.Contains(r.SetIdentifier, newClusterSuffix) && weightPercentage != 100 {
			resourceRecordSetInput := a.buildChangeTrafficWeightsInput(r.Name, r.SetIdentifier, weight, recordType, newClusterSuffix)
			fmt.Printf("Switching %v percent of traffic from %s cluster to %s cluster - %s Type record\n", weightPercentage, oldClusterSuffix, r.SetIdentifier, r.Type)

			resp, err := a.route53Client.ChangeResourceRecordSets(context.TODO(), resourceRecordSetInput)

			if err != nil {
				return errors.Wrapf(err, "Failed to update the %s type DNS records", r.Type)
			}
			fmt.Printf("Successfully processed Route53 change: %s", *resp.ChangeInfo.Id)
		} else if r.Type == r53types.RRType(recordType) && strings.Contains(r.SetIdentifier, oldClusterSuffix) && weightPercentage == 100 {
			resourceRecordSetInput := a.buildChangeTrafficWeightsInput(r.Name, r.SetIdentifier, weight, recordType, oldClusterSuffix)
			fmt.Printf("Switching %v percent of traffic from %s cluster to %s cluster - %s Type record\n", weightPercentage, oldClusterSuffix, r.SetIdentifier, r.Type)

			resp, err := a.route53Client.ChangeResourceRecordSets(context.TODO(), resourceRecordSetInput)

			if err != nil {
				return errors.Wrapf(err, "Failed to update the %s type DNS records", r.Type)
			}
			fmt.Printf("Successfully processed Route53 change: %s", *resp.ChangeInfo.Id)
		}
	}
	return nil
}

func (a *App) buildChangeTrafficWeightsInput(zoneName string, identifier string, weight int64, recordType string, clusterName string) *route53.ChangeResourceRecordSetsInput {
	aliasTarget := r53types.AliasTarget{DNSName: aws.String(fmt.Sprintf("%s-%s", clusterName, zoneName)), HostedZoneId: &a.config.HostedZoneID, EvaluateTargetHealth: false}
	record := r53types.ResourceRecordSet{Name: &zoneName, SetIdentifier: &identifier, Weight: &weight, Type: r53types.RRType(recordType), AliasTarget: &aliasTarget}
	changeInput := r53types.Change{Action: "UPSERT", ResourceRecordSet: &record}

	changeInputArray := make([]r53types.Change, 0)
	changeInputArray = append(changeInputArray, changeInput)

	changeSet := r53types.ChangeBatch{Changes: changeInputArray}
	changeWeightInput := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch:  &changeSet,
		HostedZoneId: &a.config.HostedZoneID,
	}
	fmt.Println(changeWeightInput)
	return changeWeightInput
}

func convertPerecentageToWeight(trafficSwitchPercentage int64) (int64, error) {
	var trafficWeight int64
	switch trafficSwitchPercentage {
	case 10:
		trafficWeight = 30
		return trafficWeight, nil
	case 50:
		trafficWeight = 255
		return trafficWeight, nil
	case 100:
		trafficWeight = 0
		return trafficWeight, nil
	}
	return 0, errors.New("Traffic switch can only be 10, 50 or 100 percent")
}

func validateClusterNameInputs(records []recordSetInfo, clusterSuffix string) bool {
	for _, r := range records {
		if strings.Contains(r.SetIdentifier, clusterSuffix) {
			return true
		}
	}
	return false
}
