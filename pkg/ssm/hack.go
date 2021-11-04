package ssm

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// This file contains some AWS SDK/SSM structs that are not generated in public AWS SDK.
// They are mostly taken from amazon-ssm-agent project
// https://github.com/aws/amazon-ssm-agent/blob/mainline/vendor/github.com/aws/aws-sdk-go/service/ssm/api.go

const (
	operationRegisterManagedInstance = "RegisterManagedInstance"
	methodPost                       = "POST"
)

type AWSRequester interface {
	NewRequest(operation *request.Operation, params interface{}, data interface{}) AWSRequest
}

type AWSRequest interface {
	Send() error
}

type awsClientAdapter struct {
	*ssm.SSM
}

func (a *awsClientAdapter) NewRequest(operation *request.Operation,
	params interface{}, data interface{}) AWSRequest {
	return a.SSM.NewRequest(operation, params, data)
}

type registerManagedInstanceInput struct {
	_ struct{} `type:"structure"`

	// ActivationCode is a required field
	ActivationCode *string `min:"20" type:"string" required:"true"`

	//lint:ignore ST1003 used by AWS SDK
	// ActivationId is a required field
	ActivationId *string `min:"36" type:"string" required:"true"`

	// Fingerprint is a required field
	Fingerprint *string `type:"string" required:"true"`

	// PublicKey is a required field
	PublicKey *string `min:"392" type:"string" required:"true"`

	PublicKeyType *string `type:"string" enum:"PublicKeyType"`
}

type registerManagedInstanceOutput struct {
	_ struct{} `type:"structure"`

	//lint:ignore ST1003 used by AWS SDK
	InstanceId *string `min:"10" type:"string"`
}
