// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"flag"
	"encoding/json"
	"fmt"
	"os"

	"log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	awscompat "github.com/salrashid123/awscompat/google"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"	
	"github.com/google/uuid"
)

type CredConfig struct {
	flAudience string
	flCredentialFile         string
	flAWSArn string
	flAWSSessionName string
	flAWSDuration   int
}

// https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html
type processCredentialsResponse struct {
	Version int `json:"Version"`
	AccessKeyId string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken string `json:"SessionToken"`
	Expiration string `json:"Expiration"`
}

const (
	ISO8601 = "2006-01-02T15:04:05-0700"
)
var (
	cfg = &CredConfig{}
)

// https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html

func init() {

	flag.StringVar(&cfg.flAudience, "audience", "https://sts.amazonaws.com", "(optional) audience value for the id_token")
	flag.StringVar(&cfg.flCredentialFile, "gcp-credential-file", "", "(optional) Use GCP ServiceAccount Credential File")
	flag.StringVar(&cfg.flAWSArn, "aws-arn", "", "(required) AWS ARN Value")
	flag.StringVar(&cfg.flAWSSessionName, "aws-session-name", fmt.Sprintf("gcp-%s", uuid.New().String()), "AWS SessionName")
	flag.IntVar(&cfg.flAWSDuration, "aws-duration", 3600, "STS Token Duration")

	flag.Parse()

	argError := func(s string, v ...interface{}) {
		//flag.PrintDefaults()
		log.Fatalf("Invalid Argument error: "+s, v...)
		os.Exit(1)
	}

	if cfg.flAWSArn == "" {		
		argError("-aws-arn cannot be null")		
	}
}

func main() {
	ctx := context.Background()
	var ts oauth2.TokenSource
	var err error

	if cfg.flCredentialFile != "" {
		ts, err = idtoken.NewTokenSource(ctx, cfg.flAudience, idtoken.WithCredentialsFile(cfg.flCredentialFile))
	} else {
	  	ts, err = idtoken.NewTokenSource(ctx,  cfg.flAudience)
	}
	if err != nil {
		log.Fatalf("Error creating google TokenSource: %v", err)
	}

	creds, err := awscompat.NewGCPAWSCredentials(ts, &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:         aws.String(cfg.flAWSArn),
		RoleSessionName: aws.String(cfg.flAWSSessionName),
	})
	if err != nil {
		log.Fatalf("Error creating STS Credential  %v", err)
	}

	val, err := creds.Get()
	if err != nil {
		log.Fatalf("Error parsing STS Credentials %v", err)
	}

	t, err := creds.ExpiresAt()
	if err != nil {
		log.Fatalf("Error getting Expiration Time %v", err)
	}

	resp := &processCredentialsResponse{
		Version: 1,
		AccessKeyId: val.AccessKeyID,
		SecretAccessKey: val.SecretAccessKey,
		SessionToken: val.SessionToken,
		Expiration:  fmt.Sprintf("%s",t.Format(ISO8601)),
	}

	m, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error marshalling processCredential output %v", err)
	}
	fmt.Println(string(m))
}