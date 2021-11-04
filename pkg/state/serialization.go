package state

const (
	DirSsmVault             = "/var/lib/amazon/ssm/Vault"
	FileManifest            = "Manifest"
	FileRegistrationKey     = "Store/RegistrationKey"
	FileInstanceFingerprint = "Store/InstanceFingerprint"
	EksConnectorConfig      = "EksConnectorConfig"
)

type SerializedState map[string]string

type manifestState struct {
	InstanceFingerprint string `json:"InstanceFingerprint"`
	RegistrationKey     string `json:"RegistrationKey"`
}

// instanceFingerprintState file /var/lib/amazon/ssm/Vault/Store/InstanceFingerprint
// https://github.com/aws/amazon-ssm-agent/blob/a62919edde2dccc0b84044d55d1b863ebc7baf00/agent/managedInstances/fingerprint/fingerprint.go#L37-L41
type instanceFingerprintState struct {
	Fingerprint         string            `json:"fingerprint"`
	HardwareHash        map[string]string `json:"hardwareHash"`
	SimilarityThreshold int               `json:"similarityThreshold"`
}

// registrationKeyState file /var/lib/amazon/ssm/Vault/Store/RegistrationKey
// https://github.com/aws/amazon-ssm-agent/blob/897de484b0bb35d6327d66d9f5d7308b3585d7cc/agent/managedInstances/registration/instance_info.go#L31-L39
type registrationKeyState struct {
	InstanceID            string `json:"instanceID"`
	Region                string `json:"region"`
	InstanceType          string `json:"InstanceType"`
	AvailabilityZone      string `json:"availabilityZone"`
	PrivateKey            string `json:"privateKey"`
	PrivateKeyType        string `json:"privateKeyType"`
	PrivateKeyCreatedDate string `json:"privateKeyCreatedDate"`
}

// eksConnectorConfig is used to capture the original eks connector configuration
type eksConnectorConfig struct {
	ActivationId string `json:"activationId"`
}
