package google

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"

	creds "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"golang.org/x/oauth2"
)

const (
	GCPProviderName  = "GCPProvider"
	refreshTolerance = 60
)

type GCPProvider struct {
	identityInput sts.AssumeRoleWithWebIdentityInput
	tokenSource   oauth2.TokenSource
	expiry        *time.Time
}

func NewGCPAWSCredentials(ts oauth2.TokenSource, cfg *sts.AssumeRoleWithWebIdentityInput) (creds.Credentials, error) {
	return *creds.NewCredentials(&GCPProvider{identityInput: *cfg, tokenSource: ts}), nil
}

func (s *GCPProvider) Retrieve() (creds.Value, error) {
	tok, err := s.tokenSource.Token()
	if err != nil {
		return creds.Value{}, err
	}
	s.identityInput.WebIdentityToken = &tok.AccessToken

	sess, err := session.NewSession(&aws.Config{})
	svc := sts.New(sess)
	result, err := svc.AssumeRoleWithWebIdentity(&s.identityInput)
	if err != nil {
		return creds.Value{}, err
	}

	v := creds.Value{
		AccessKeyID:     aws.StringValue(result.Credentials.AccessKeyId),
		SecretAccessKey: aws.StringValue(result.Credentials.SecretAccessKey),
		SessionToken:    aws.StringValue(result.Credentials.SessionToken),
	}
	if v.ProviderName == "" {
		v.ProviderName = GCPProviderName
	}
	s.expiry = result.Credentials.Expiration

	return v, nil
}

func (s *GCPProvider) IsExpired() bool {
	if time.Now().Add(time.Second * time.Duration(refreshTolerance)).After(*s.expiry) {
		return true
	}
	return false
}
