package asyncapi

// This file is automatically generated, please do not edit.

const Version = "v0.0.2"

// AssetPayload represents a AssetPayload model.
type AssetPayload struct {
	Id          string
	Team        *Team
	Alias       string
	Rolfp       string
	Scannable   bool
	AssetType   *AssetType
	Identifier  string
	Annotations []*Annotation
}

// Team represents a Team model.
type Team struct {
	Id          string
	Name        string
	Description string
	Tag         string
}

// AssetType represents an enum of string.
type AssetType string

const (
	AssetTypeIp            AssetType = "IP"
	AssetTypeDomainName              = "DomainName"
	AssetTypeHostname                = "Hostname"
	AssetTypeAwsAccount              = "AWSAccount"
	AssetTypeIpRange                 = "IPRange"
	AssetTypeDockerImage             = "DockerImage"
	AssetTypeWebAddress              = "WebAddress"
	AssetTypeGitRepository           = "GitRepository"
	AssetTypeGcpProject              = "GCPProject"
)

// Annotation represents a Annotation model.
type Annotation struct {
	Key   string
	Value string
}
