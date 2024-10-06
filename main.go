package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func main() {
	// Select AWS profile
	profile, err := promptForAWSProfile()
	if err != nil {
		log.Fatalf("Failed to select AWS profile: %v", err)
	}
	fmt.Printf("Selected AWS profile: %s\n", profile)

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile),
		config.WithRegion("us-east-1"), // Add default region
	)
	if err != nil {
		log.Fatalf("Unable to load AWS configuration: %v", err)
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Get all regions
	loadingDone = make(chan struct{})
	loadingMessage := "Retrieving all AWS regions..."
	go printWithLoading(&loadingMessage)

	regionsChan := make(chan []string)
	errorChan := make(chan error)

	go func() {
		regions, err := getAllRegions(ec2Client)
		if err != nil {
			errorChan <- err
			return
		}
		regionsChan <- regions
	}()

	var regions []string
	select {
	case regions = <-regionsChan:
		close(loadingDone)
		<-time.After(100 * time.Millisecond) // Wait for the last frame to be printed
		printComplete()
		fmt.Printf("Retrieved %d regions\n", len(regions))
	case err := <-errorChan:
		close(loadingDone)
		printError()
		log.Fatalf("Failed to get region list: %v", err)
	case <-time.After(30 * time.Second):
		close(loadingDone)
		printError()
		log.Fatalf("Timeout while retrieving regions")
	}

	// Get all regions with EC2 instances
	loadingDone = make(chan struct{})
	loadingMessage = "Checking EC2 instances in each region..."
	go printWithLoading(&loadingMessage)
	startTime := time.Now()
	regionsWithInstances, err := getRegionsWithInstances(cfg, regions)
	endTime := time.Now()
	close(loadingDone)
	<-time.After(100 * time.Millisecond) // Wait for the last frame to be printed
	printComplete()
	fmt.Printf("Time taken to check regions: %v\n", endTime.Sub(startTime))

	if err != nil {
		log.Printf("Warning: Error occurred while checking regions: %v", err)
	}

	if len(regionsWithInstances) == 0 {
		fmt.Println("No regions with EC2 instances found. Exiting program.")
		return
	}

	fmt.Printf("Found %d regions with EC2 instances\n", len(regionsWithInstances))

	// Let user select a region
	region, err := promptForRegion(regionsWithInstances)
	if err != nil {
		log.Fatalf("Failed to select region: %v", err)
	}

	var allInstances []types.Instance
	var instanceRegions []string

	if region == "View all regions" {
		fmt.Println("Retrieving EC2 instances from all regions...")
		for _, r := range regionsWithInstances {
			cfg.Region = r
			ec2Client := ec2.NewFromConfig(cfg)
			instances, err := getEC2Instances(ec2Client)
			if err != nil {
				log.Printf("Warning: Failed to get EC2 instances for region %s: %v", r, err)
				continue
			}
			allInstances = append(allInstances, instances...)
			instanceRegions = append(instanceRegions, make([]string, len(instances))...)
			for i := range instances {
				instanceRegions[len(instanceRegions)-len(instances)+i] = r
			}
		}
	} else {
		// Use the selected region to create a new EC2 client
		cfg.Region = region
		ec2Client := ec2.NewFromConfig(cfg)

		// Get EC2 instance information
		allInstances, err = getEC2Instances(ec2Client)
		if err != nil {
			log.Fatalf("Failed to get EC2 instance list: %v", err)
		}
		instanceRegions = make([]string, len(allInstances))
		for i := range allInstances {
			instanceRegions[i] = region
		}
	}

	// Display EC2 instance information
	displayEC2InstancesTable(cfg, allInstances, instanceRegions)
}
