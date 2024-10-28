package main

import (
	"context"
	"fmt"
	"time"

	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	awscompat "github.com/salrashid123/awscompat/google"
	"google.golang.org/api/idtoken"
)

const ()

func main() {

	aud := "https://foo.bar"
	jsonCert := "/path/to/svc.json"

	ctx := context.Background()
	ts, err := idtoken.NewTokenSource(ctx, aud, idtoken.WithCredentialsFile(jsonCert))
	// or on GCE/GKE/Run/GCF, omit the certificate file
	//ts, err := idtoken.NewTokenSource(ctx, aud)
	if err != nil {
		log.Fatalf("unable to create TokenSource: %v", err)
	}

	region := "us-east-2"
	creds, err := awscompat.NewGCPAWSCredentials(ts, region, &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:         aws.String("arn:aws:iam::291738886548:role/s3role"),
		RoleSessionName: aws.String("app1"),
	})
	if err != nil {
		log.Fatalf("Error creating Credentials  %v", err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region), config.WithCredentialsProvider(creds))

	s3client := s3.NewFromConfig(cfg)

	sresp, err := s3client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{Bucket: aws.String("mineral-minutia")})
	if err != nil {
		log.Fatalf("Error listing objects:  %v", err)
	}

	for _, item := range sresp.Contents {
		fmt.Printf("Name %v  %v\n:", time.Now(), *item.Key)
	}

}
