package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/manifoldco/promptui"
	"gopkg.in/ini.v1"
)

func promptForAWSProfile() (string, error) {
	profiles, err := getAWSProfiles()
	if err != nil {
		return "", err
	}

	if len(profiles) == 0 {
		return "", fmt.Errorf("no AWS profiles found, please ensure you have configured AWS CLI")
	}

	prompt := promptui.Select{
		Label: "Select AWS Profile",
		Items: profiles,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select AWS profile: %v", err)
	}

	return result, nil
}

func getAWSProfiles() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	credentialsPath := filepath.Join(homeDir, ".aws", "credentials")
	cfg, err := ini.Load(credentialsPath)
	if err != nil {
		return nil, err
	}

	var profiles []string
	for _, section := range cfg.Sections() {
		if section.Name() != "DEFAULT" {
			profiles = append(profiles, section.Name())
		}
	}

	return profiles, nil
}

func promptForRegion(regions []string) (string, error) {
	allRegionsOption := "View all regions"
	options := append([]string{allRegionsOption}, regions...)

	prompt := promptui.Select{
		Label: "Select AWS Region",
		Items: options,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select region: %v", err)
	}

	return result, nil
}

func getAllRegions(client *ec2.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	input := &ec2.DescribeRegionsInput{}
	result, err := client.DescribeRegions(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %w", err)
	}

	var regions []string
	for _, region := range result.Regions {
		regions = append(regions, *region.RegionName)
	}

	return regions, nil
}
