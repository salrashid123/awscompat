package google

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	creds "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"golang.org/x/oauth2"
)

const (
	GCPProviderName = "GCPProvider"
)

type GCPProvider struct {
	identityInput sts.AssumeRoleWithWebIdentityInput
	tokenSource   oauth2.TokenSource
	region        string
}

func NewGCPAWSCredentials(ts oauth2.TokenSource, region string, cfg *sts.AssumeRoleWithWebIdentityInput) (creds.CredentialsProvider, error) {
	// todo, validate input parameters...or just fail on retrieve()
	return &GCPProvider{identityInput: *cfg, region: region, tokenSource: ts}, nil
}

func (s *GCPProvider) Retrieve(context.Context) (creds.Credentials, error) {

	tok, err := s.tokenSource.Token()
	if err != nil {
		return creds.Credentials{}, err
	}

	s.identityInput.WebIdentityToken = &tok.AccessToken

	c, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(s.region))
	if err != nil {
		return creds.Credentials{}, err
	}
	client := sts.NewFromConfig(c)
	stsOutput, err := client.AssumeRoleWithWebIdentity(context.Background(), &s.identityInput)
	if err != nil {
		return creds.Credentials{}, err
	}

	return creds.Credentials{
		AccessKeyID:     aws.ToString(stsOutput.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(stsOutput.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(stsOutput.Credentials.SessionToken),
		Source:          GCPProviderName,
		CanExpire:       true,
		Expires:         *stsOutput.Credentials.Expiration,
	}, nil

}
