package types

type Cluster struct {
	MetaData *ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Name is the name of this cluster
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Version is the k8s version of this cluster
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// Ingress is the ingress configuration for services in this cluster
	Ingress *Ingress `json:"ingress,omitempty" yaml:"ingress,omitempty"`
}

type Ingress struct {
	// URLFormat is the format string for generating URL's to different services in the cluster
	URLFormat string `json:"urlFormat,omitempty" yaml:"urlFormat,omitempty"`
}
