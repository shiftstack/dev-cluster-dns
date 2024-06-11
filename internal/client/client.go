package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

const (
	defaultAWSRegion = "us-east-1"
)

func GetRoute53Client(ctx context.Context, profile string) (*route53.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile), config.WithDefaultRegion(defaultAWSRegion))
	if err != nil {
		return nil, err
	}

	return route53.NewFromConfig(cfg), nil
}
