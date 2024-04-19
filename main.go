package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"log"
	"time"
)

type StateChangeNotification struct {
	Version    string    `json:"version"`
	Id         string    `json:"id"`
	DetailType string    `json:"detail-type"`
	Source     string    `json:"source"`
	Account    string    `json:"account"`
	Time       time.Time `json:"time"`
	Region     string    `json:"region"`
	Resources  []string  `json:"resources"`
	Detail     struct {
		InstanceId string `json:"instance-id"`
		State      string `json:"state"`
	} `json:"detail"`
}

func HandleRequest(ctx context.Context, event *StateChangeNotification) (*string, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}
	message := fmt.Sprintf("Instance %s: %s", event.Detail.InstanceId, event.Detail.State)

	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
	}

	e := ec2.NewFromConfig(sdkConfig)
	params2 := ec2.DescribeInstancesInput{
		DryRun:      nil,
		Filters:     nil,
		InstanceIds: []string{event.Detail.InstanceId},
		MaxResults:  nil,
		NextToken:   nil,
	}
	result, err2 := e.DescribeInstances(ctx, &params2)
	if err2 != nil {
		log.Fatal(err2)
	} else {
		instance := result.Reservations[0].Instances[0]
		networkAssociation := instance.NetworkInterfaces[0].Association
		if networkAssociation != nil {
			publicIp := *networkAssociation.PublicIp

			tags := instance.Tags

			var domainName string
			var hostedZoneId string

			for _, tag := range tags {
				if *tag.Key == "DomainName" {
					log.Printf("Tag 'DomainName' value = %s", *tag.Value)
					domainName = *tag.Value
				} else if *tag.Key == "HostedZoneId" {
					log.Printf("Tag 'HostedZoneId' value = %s", *tag.Value)
					hostedZoneId = *tag.Value
				}
			}

			setRecord(ctx, hostedZoneId, domainName, publicIp)
			log.Println(hostedZoneId)
			log.Println(domainName)
			log.Println(publicIp)
		}
	}

	return &message, nil
}

func setRecord(ctx context.Context, hostedZoneId string, domainName string, ipAddr string) {
	params := route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(domainName),
						Type: types.RRTypeA,
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String(ipAddr)},
						},
						TTL: aws.Int64(300),
					},
				},
			},
		},
		HostedZoneId: aws.String(hostedZoneId),
	}

	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return
	}

	r := route53.NewFromConfig(sdkConfig)
	res, err := r.ChangeResourceRecordSets(ctx, &params)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(res)
	}
}

func main() {
	lambda.Start(HandleRequest)
}
