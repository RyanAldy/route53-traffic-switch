package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	oldClusterSuffix := flag.String("old", "foxdev", "Suffix of the old cluster")
	newClusterSuffix := flag.String("new", "foxdev", "Suffix of the new cluster")
	trafficSwitchPercentage := flag.Int64("traffic", 10, "Percentage of traffic to switch to new cluster")
	region := flag.String("region", "euc1", "Short region name. Can be euc1, use1 or apn1")
	environment := flag.String("environment", "internal-dev", "Environment to run this against")
	flag.Parse()

	config, err := New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	a, err := NewApp(*config, oldClusterSuffix, newClusterSuffix, trafficSwitchPercentage, region, environment)

	if err != nil {
		fmt.Printf("Error running application due to %s", err)
	}

	message, err := a.handler()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(message)
}
