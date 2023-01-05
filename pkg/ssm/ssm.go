// Package ssm provides abstracted interaction to AWS SSM service
package ssm

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"k8s.io/klog/v2"

	"github.com/aws/amazon-eks-connector/pkg/config"
)

// mostly "borrowed" from
// https://github.com/aws/amazon-ssm-agent/blob/mainline/agent/ssm/anonauth/anon_service.go

// Client is an interface to the operations of the SSM service.
type Client interface {
	RegisterManagedInstance(activationID, activationCode, publicKey, publicKeyType, fingerprint string) (string, error)
	Region() string
}

// sdkClient is an service wrapper that delegates to the ssm sdk.
type sdkClient struct {
	agentConfig *config.AgentConfig
	sdk         AWSRequester
}

// NewClient creates a new SSM client instance.
func NewClient(agentConfig *config.AgentConfig) Client {
	awsConfig := aws.NewConfig().
		WithRegion(agentConfig.Region).
		WithCredentials(credentials.AnonymousCredentials)

	if agentConfig.Endpoint != "" {
		klog.Infof("overriding SSM endpoint to %s", agentConfig.Endpoint)
		awsConfig.Endpoint = &agentConfig.Endpoint
	}

	// Create a session to share service client config and handlers with
	ssmSess := session.Must(session.NewSession(awsConfig))
	ssmService := ssm.New(ssmSess)
	return &sdkClient{
		agentConfig: agentConfig,
		sdk:         &awsClientAdapter{ssmService},
	}
}

// RegisterManagedInstance calls the RegisterManagedInstance SSM API.
// It is not included in public AWS SDK, so we are doing it in the hard way.
func (svc *sdkClient) RegisterManagedInstance(activationID, activationCode, publicKey, publicKeyType, fingerprint string) (string, error) {
	op := &request.Operation{
		Name:       operationRegisterManagedInstance,
		HTTPMethod: methodPost,
		HTTPPath:   "/",
	}

	params := &registerManagedInstanceInput{
		ActivationId:   aws.String(activationID),
		ActivationCode: aws.String(activationCode),
		PublicKey:      aws.String(publicKey),
		PublicKeyType:  aws.String(publicKeyType),
		Fingerprint:    aws.String(fingerprint),
	}

	output := &registerManagedInstanceOutput{}

	req := svc.sdk.NewRequest(op, params, output)

	if err := req.Send(); err != nil {
		return "", err
	}
	return *output.InstanceId, nil
}

func (svc *sdkClient) Region() string {
	return svc.agentConfig.Region
}
