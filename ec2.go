package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/olekukonko/tablewriter"
	"os"
	"sync"
	"time"
)

func getEC2Instances(client *ec2.Client) ([]types.Instance, error) {
	input := &ec2.DescribeInstancesInput{}
	result, err := client.DescribeInstances(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var instances []types.Instance
	for _, reservation := range result.Reservations {
		instances = append(instances, reservation.Instances...)
	}

	return instances, nil
}

func displayEC2InstancesTable(cfg aws.Config, instances []types.Instance, regions []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Region", "Instance Name", "Instance ID", "Instance Type", "State", "Launch Time", "Total Runtime", "Estimated Hourly Cost", "Total Cost"})
	table.SetAutoMergeCells(false)
	table.SetRowLine(true)

	// Set only the region column to auto-merge
	table.SetAutoMergeCellsByColumnIndex([]int{0})

	var totalHourlyCost, totalCost float64

	for i, instance := range instances {
		name := getInstanceName(instance)
		launchTime := instance.LaunchTime.Format("2006-01-02 15:04:05")
		runningTime := calculateRunningTime(*instance.LaunchTime)
		hourlyCost := estimateInstanceCost(cfg, instance, regions[i])
		instanceTotalCost := calculateTotalCost(hourlyCost, *instance.LaunchTime)

		totalHourlyCost += hourlyCost
		totalCost += instanceTotalCost

		table.Append([]string{
			regions[i],
			name,
			*instance.InstanceId,
			string(instance.InstanceType),
			string(instance.State.Name),
			launchTime,
			runningTime,
			fmt.Sprintf("$%.4f", hourlyCost),
			fmt.Sprintf("$%.4f", instanceTotalCost),
		})
	}

	table.Render()

	// Output total information separately
	fmt.Printf("\nTotal:\n")
	fmt.Printf("Estimated Total Hourly Cost: $%.4f\n", totalHourlyCost)
	fmt.Printf("Total Cost: $%.4f\n", totalCost)
}

func getRegionsWithInstances(cfg aws.Config, regions []string) ([]string, error) {
	var wg sync.WaitGroup
	regionsWithInstances := make(chan string, len(regions))
	errorChan := make(chan error, len(regions))
	semaphore := make(chan struct{}, 10) // Limit concurrency to 10

	for _, region := range regions {
		wg.Add(1)
		go func(regionName string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			// Create new configuration with only the region changed
			regionalCfg := cfg.Copy()
			regionalCfg.Region = regionName
			regionalClient := ec2.NewFromConfig(regionalCfg)

			instances, err := getEC2Instances(regionalClient)
			if err != nil {
				errorChan <- fmt.Errorf("failed to get EC2 instances for region %s: %v", regionName, err)
				return
			}
			if len(instances) > 0 {
				regionsWithInstances <- regionName
			}
		}(region)
	}

	go func() {
		wg.Wait()
		close(regionsWithInstances)
		close(errorChan)
	}()

	var result []string
	for region := range regionsWithInstances {
		result = append(result, region)
	}

	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return result, fmt.Errorf("errors occurred while getting region information: %v", errors)
	}

	return result, nil
}

func getInstanceName(instance types.Instance) string {
	for _, tag := range instance.Tags {
		if *tag.Key == "Name" {
			return *tag.Value
		}
	}
	return "N/A"
}

func calculateRunningTime(launchTime time.Time) string {
	duration := time.Since(launchTime)
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}
