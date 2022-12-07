package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53"

	r53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func (a *App) handler() (string, error) {

	ctx := context.TODO()
	oldClusterSuffix := flag.String("old", "foxdev", "Suffix of the old cluster")
	newClusterSuffix := flag.String("new", "foxdev", "Suffix of the new cluster")
	trafficSwitchPercentage := flag.Int64("traffic", 10, "Percentage of traffic to switch to new cluster")
	region := flag.String("region", "euc1", "Short region name. Can be euc1, use1 or apn1")
	environment := flag.String("environment", "internal-dev", "Environment to run this against")
	flag.Parse()

	trafficWeight := convertPerecentageToWeight(*trafficSwitchPercentage)

	// To be parameterised also - will be different for prod
	dnsInput := "internal-dev.dazn-gateway.com"

	hostedZoneInput := &route53.ListHostedZonesByNameInput{
		DNSName: &dnsInput,
	}

	hostedZonesResult, err := a.route53Client.ListHostedZonesByName(ctx, hostedZoneInput)
	if err != nil {
		fmt.Println(err)
	}

	extractedInfo := hostedZonesResult.HostedZones

	zoneInfo := []hostedZoneInfo{}
	for _, info := range extractedInfo {
		zoneInfo = append(zoneInfo, hostedZoneInfo{Id: *info.Id, Name: *info.Name})
	}

	// Loop over struct to check which one matches the Name we're looking for
	//  I need to get the id of the Hosted Zone to get more info
	var hostedZoneId string
	for _, zone := range zoneInfo {
		if zone.Name == fmt.Sprintf("%s.dazn-gateway.com.", *environment) {
			hostedZoneId = zone.Id
		}
	}
	// fmt.Println(hostedZoneId)

	idInput := &route53.ListResourceRecordSetsInput{
		HostedZoneId: &hostedZoneId,
	}

	recordSets, err := a.route53Client.ListResourceRecordSets(ctx, idInput)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	recordInfo := []recordSetInfo{}
	for _, record := range recordSets.ResourceRecordSets {
		if *record.Name == fmt.Sprintf("%s.%s.dazn-gateway.com.", *region, *environment) {
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
			HostedZoneId:          &hostedZoneId,
			StartRecordName:       recordSets.NextRecordName,
			StartRecordType:       recordSets.NextRecordType,
			StartRecordIdentifier: recordSets.NextRecordIdentifier,
		}
		newRecordSets, _ := a.route53Client.ListResourceRecordSets(ctx, paginateIdInput)
		for paginate {
			for _, newRecord := range newRecordSets.ResourceRecordSets {
				if *newRecord.Name == fmt.Sprintf("%s.%s.dazn-gateway.com.", *region, *environment) {
					recordInfo = append(recordInfo, recordSetInfo{
						Name:          *newRecord.Name,
						Type:          newRecord.Type,
						SetIdentifier: *newRecord.SetIdentifier,
						Weight:        *newRecord.Weight})
				}
			}

			paginate = newRecordSets.IsTruncated
			paginateIdInput := &route53.ListResourceRecordSetsInput{
				HostedZoneId:          &hostedZoneId,
				StartRecordName:       recordSets.NextRecordName,
				StartRecordType:       recordSets.NextRecordType,
				StartRecordIdentifier: recordSets.NextRecordIdentifier,
			}
			newRecordSets, _ = a.route53Client.ListResourceRecordSets(ctx, paginateIdInput)
		}
	}

	// For Debugging only
	// fmt.Println("Record Info: ", recordInfo)

	a.switchTraffic(recordInfo, *oldClusterSuffix, *newClusterSuffix, trafficWeight, *trafficSwitchPercentage, "A")
	a.switchTraffic(recordInfo, *oldClusterSuffix, *newClusterSuffix, trafficWeight, *trafficSwitchPercentage, "AAAA")
	return "Successfully switched over traffic", nil
}

func (a *App) switchTraffic(records []recordSetInfo, oldClusterSuffix string, newClusterSuffix string, weight int64, weightPercentage int64, recordType string) error {
	var dnsChanges int = 0

	for _, r := range records {
		if r.Type == r53types.RRType(recordType) && strings.Contains(r.SetIdentifier, newClusterSuffix) {
			// only need to use old cluster suffix when 100% switchover
			buildChangeTrafficWeightsInput(r.Name, r.SetIdentifier, weight)
			fmt.Printf("Switching %v percent of traffic from %s cluster to %s cluster - %s Type record\n", weightPercentage, oldClusterSuffix, r.SetIdentifier, r.Type)
			// resp, err := a.route53Client.ChangeResourceRecordSets(context.TODO(), resourceRecordSetInput)

			// if err != nil {
			// 	return errors.Wrapf(err, "failed to update the %s type DNS records", r.Type)
			// }
			dnsChanges++
		}
	}
	// Amend this logic?
	if dnsChanges == 0 {
		fmt.Println("Cluster DNS record does not exist - no changes were made")
		os.Exit(1)
	}
	return nil
}

func buildChangeTrafficWeightsInput(zoneName string, identifier string, weight int64) *route53.ChangeResourceRecordSetsInput {
	record := r53types.ResourceRecordSet{Name: &zoneName, SetIdentifier: &identifier, Weight: &weight}
	changeInput := r53types.Change{Action: "UPSERT", ResourceRecordSet: &record}

	changeInputArray := make([]r53types.Change, 0)
	changeInputArray = append(changeInputArray, changeInput)

	changeSet := r53types.ChangeBatch{Changes: changeInputArray}
	changeWeightInput := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &changeSet,
	}
	return changeWeightInput
}

func convertPerecentageToWeight(trafficSwitchPercentage int64) int64 {
	var trafficWeight int64
	switch trafficSwitchPercentage {
	case 10:
		trafficWeight = 30
	case 50:
		trafficWeight = 255
	}
	return trafficWeight
}
