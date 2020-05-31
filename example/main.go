package main

import (
	"context"
	"fmt"

	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/salrashid123/awscompat/google"

	"google.golang.org/api/idtoken"
)

const ()

// https://pkg.go.dev/google.golang.org/api@v0.23.0/idtoken

func main() {

	aud := "https://foo.bar"
	jsonCert := "/path/to/svc.json"

	ctx := context.Background()

	ts, err := idtoken.NewTokenSource(ctx, aud, idtoken.WithCredentialsFile(jsonCert))
	if err != nil {
		log.Fatalf("unable to create TokenSource: %v", err)
	}

	creds, err := google.NewGCPAWSCredentials(ts, &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:         aws.String("arn:aws:iam::291738886548:role/s3webreaderrole"),
		RoleSessionName: aws.String("app1"),
	})
	if err != nil {
		log.Fatalf("Error creatint Credentials  %v", err)
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials: &creds,
		Region:      aws.String("us-east-2")},
	)
	svcs3 := s3.New(sess)

	sresp, err := svcs3.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String("mineral-minutia")})
	if err != nil {
		log.Fatalf("Error listing objects:  %v", err)
	}

	for _, item := range sresp.Contents {
		fmt.Println("Name:         ", *item.Key)
	}

}
