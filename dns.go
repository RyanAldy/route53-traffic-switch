package main

// Parameters needed for this to work:
// 1) suffix of old cluster
// 2) suffix of the new cluster
// 3) Percentage of Traffic switch

import (
	"flag"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type hostedZoneInfo struct {
	Id   string
	Name string
}

type recordSetInfo struct {
	Name           string
	Type           string
	SetIndentifier string
	Weight         int64
}

func main() {
	// Need to export AWS env vars for this to run at the moment

	oldClusterSuffix := flag.String("old", "foxdev", "Suffix of the old cluster")
	NewClusterSuffix := flag.String("new", "foxdev", "Suffix of the new cluster")
	trafficSwitchPercentage := flag.Int64("traffic", 10, "Percentage of traffic to switch to new cluster")
	region := flag.String("region", "euc1", "Short region name. Can be euc1, use1 or apn1")
	environment := flag.String("environment", "internal-dev", "Environment to run this against")
	flag.Parse()

	trafficWeight := convertPerecentageToWeight(*trafficSwitchPercentage)

	// For stage-dev this would be: dev.dazn-gateway.com
	// To be parameterised also?
	// dnsInput := "dev.dazn-gateway.com"
	dnsInput := "internal-dev.dazn-gateway.com"
	svc := route53.New(session.New())

	hostedZoneInput := &route53.ListHostedZonesByNameInput{
		DNSName: &dnsInput,
	}

	hostedZonesResult, err := svc.ListHostedZonesByName(hostedZoneInput)
	if err != nil {
		fmt.Println(err)
	}

	extractedInfo := hostedZonesResult.HostedZones

	// Save result into struct
	zoneInfo := []hostedZoneInfo{}
	for _, info := range extractedInfo {
		zoneInfo = append(zoneInfo, hostedZoneInfo{Id: *info.Id, Name: *info.Name})
	}

	// Loop over struct to check which one matches the Name we're looking for
	//  I need to get the id of the Hosted Zone to get more info
	// Record.name will be dev.dazn-gateway.com was gateway.transit.dazn-dev.com.
	var hostedZoneId string
	for _, zone := range zoneInfo {
		if zone.Name == fmt.Sprintf("%s.dazn-gateway.com.", *environment) {
			hostedZoneId = zone.Id
		}
	}
	fmt.Println(hostedZoneId)

	idInput := &route53.ListResourceRecordSetsInput{
		HostedZoneId: &hostedZoneId,
	}

	recordSets, err := svc.ListResourceRecordSets(idInput)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	recordInfo := []recordSetInfo{}
	for _, record := range recordSets.ResourceRecordSets {
		if *record.Name == fmt.Sprintf("%s.%s.dazn-gateway.com.", *region, *environment) {
			recordInfo = append(recordInfo, recordSetInfo{
				Name:           *record.Name,
				Type:           *record.Type,
				SetIndentifier: *record.SetIdentifier,
				Weight:         *record.Weight})
		}
	}
	// Check if 2nd page, if so initalise vars for call
	startPagiante := recordSets.IsTruncated

	if *startPagiante {
		var paginate = true
		paginateIdInput := &route53.ListResourceRecordSetsInput{
			HostedZoneId:          &hostedZoneId,
			StartRecordName:       recordSets.NextRecordName,
			StartRecordType:       recordSets.NextRecordType,
			StartRecordIdentifier: recordSets.NextRecordIdentifier,
		}
		newRecordSets, _ := svc.ListResourceRecordSets(paginateIdInput)
		for paginate {
			for _, newRecord := range newRecordSets.ResourceRecordSets {
				if *newRecord.Name == fmt.Sprintf("%s.%s.dazn-gateway.com.", *region, *environment) {
					recordInfo = append(recordInfo, recordSetInfo{
						Name:           *newRecord.Name,
						Type:           *newRecord.Type,
						SetIndentifier: *newRecord.SetIdentifier,
						Weight:         *newRecord.Weight})
				}
			}
			// reinitalise
			// Refresh page data and next page params
			paginate = *newRecordSets.IsTruncated
			paginateIdInput := &route53.ListResourceRecordSetsInput{
				HostedZoneId:          &hostedZoneId,
				StartRecordName:       recordSets.NextRecordName,
				StartRecordType:       recordSets.NextRecordType,
				StartRecordIdentifier: recordSets.NextRecordIdentifier,
			}
			newRecordSets, _ = svc.ListResourceRecordSets(paginateIdInput)
		}
	}

	// fmt.Println("Record Info: ", recordInfo)
	// Traffic switch - put into separate function
	for _, r := range recordInfo {
		if r.Type == "A" && strings.Contains(r.SetIndentifier, *NewClusterSuffix) {
			// only need to use old cluster suffix when 100% switchover
			resourceRecordSetInput := buildChangeTrafficWeightsInput(r.Name, r.SetIndentifier, trafficWeight)
			fmt.Println(resourceRecordSetInput)
			fmt.Printf("Switching %v percent of traffic from %s cluster to %s cluster - A Type record\n", *trafficSwitchPercentage, *oldClusterSuffix, r.SetIndentifier)
			// svc.ChangeResourceRecordSets(resourceRecordSetInput)
		}
		if r.Type == "AAAA" && strings.Contains(r.SetIndentifier, *NewClusterSuffix) {
			resourceRecordSetInput := buildChangeTrafficWeightsInput(r.Name, r.SetIndentifier, trafficWeight)
			fmt.Println(resourceRecordSetInput)
			fmt.Printf("Switching %v percent of traffic from %s cluster to %s cluster - AAAA Type record\n", *trafficSwitchPercentage, *oldClusterSuffix, r.SetIndentifier)
			// svc.ChangeResourceRecordSets(resourceRecordSetInput)
		}
	}
}

func buildChangeTrafficWeightsInput(zoneName string, identifier string, weight int64) *route53.ChangeResourceRecordSetsInput {
	action := "UPSERT"
	record := &route53.ResourceRecordSet{Name: &zoneName, SetIdentifier: &identifier, Weight: &weight}
	changeInput := &route53.Change{Action: &action, ResourceRecordSet: record}

	changeInputArray := make([]*route53.Change, 0)
	changeInputArray = append(changeInputArray, changeInput)

	changeSet := &route53.ChangeBatch{Changes: changeInputArray}
	changeWeightInput := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: changeSet,
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
