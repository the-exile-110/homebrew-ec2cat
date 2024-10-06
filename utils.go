package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/fatih/color"
)

func getEC2InstancePrice(cfg aws.Config, instanceType, region string) (float64, error) {
	pricingRegion := "us-east-1"
	cfg.Region = pricingRegion
	client := pricing.NewFromConfig(cfg)

	// Build GetProducts request
	input := &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters: []types.Filter{
			{
				Field: aws.String("instanceType"),
				Type:  types.FilterTypeTermMatch,
				Value: aws.String(instanceType),
			},
			{
				Field: aws.String("regionCode"),
				Type:  types.FilterTypeTermMatch,
				Value: aws.String(region),
			},
			{
				Field: aws.String("operatingSystem"),
				Type:  types.FilterTypeTermMatch,
				Value: aws.String("Linux"),
			},
			{
				Field: aws.String("tenancy"),
				Type:  types.FilterTypeTermMatch,
				Value: aws.String("Shared"),
			},
			{
				Field: aws.String("capacitystatus"),
				Type:  types.FilterTypeTermMatch,
				Value: aws.String("Used"),
			},
		},
		MaxResults: aws.Int32(1),
	}

	// Send request
	resp, err := client.GetProducts(context.TODO(), input)
	if err != nil {
		return 0, fmt.Errorf("failed to get pricing information: %v", err)
	}

	if len(resp.PriceList) == 0 {
		return 0, fmt.Errorf("no pricing information found")
	}

	// Parse pricing information
	var priceList map[string]interface{}
	if err := json.Unmarshal([]byte(resp.PriceList[0]), &priceList); err != nil {
		return 0, fmt.Errorf("failed to parse pricing information: %v", err)
	}

	terms, ok := priceList["terms"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("failed to parse pricing information")
	}

	onDemand, ok := terms["OnDemand"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("failed to parse on-demand pricing information")
	}

	for _, v := range onDemand {
		priceDimensions, ok := v.(map[string]interface{})["priceDimensions"].(map[string]interface{})
		if !ok {
			continue
		}

		for _, pd := range priceDimensions {
			pricePerUnit, ok := pd.(map[string]interface{})["pricePerUnit"].(map[string]interface{})
			if !ok {
				continue
			}

			usd, ok := pricePerUnit["USD"].(string)
			if !ok {
				continue
			}

			price, err := strconv.ParseFloat(usd, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse price: %v", err)
			}

			return price, nil
		}
	}

	return 0, fmt.Errorf("no valid pricing information found")
}

func calculateTotalCost(hourlyCost float64, launchTime time.Time) float64 {
	runningHours := time.Since(launchTime).Hours()
	return hourlyCost * runningHours
}

func estimateInstanceCost(cfg aws.Config, instance ec2Types.Instance, region string) float64 {
	instanceType := string(instance.InstanceType)
	price, err := getEC2InstancePrice(cfg, instanceType, region)
	if err != nil {
		log.Printf("Warning: Failed to get price for instance %s: %v", *instance.InstanceId, err)
		return 0 // Return 0 if retrieval fails
	}

	// Adjust cost based on instance state
	if instance.State.Name == ec2Types.InstanceStateNameStopped {
		return 0 // Stopped instances do not incur EC2 running costs
	}

	return price
}

var loadingDone chan struct{}

func printWithLoading(message *string) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	maxLen := 0
	for {
		select {
		case <-loadingDone:
			return
		default:
			currentMessage := *message
			if len(currentMessage) > maxLen {
				maxLen = len(currentMessage)
			}
			padding := strings.Repeat(" ", maxLen-len(currentMessage))
			color.New(color.FgCyan).Printf("\r%s %s%s", frames[i], currentMessage, padding)
			time.Sleep(100 * time.Millisecond)
			i = (i + 1) % len(frames)
		}
	}
}

func printComplete() {
	color.New(color.FgGreen).Print("\r✓ ")
	fmt.Println()
}

func printError() {
	color.New(color.FgRed).Print("\r✗ ")
	fmt.Println()
}
