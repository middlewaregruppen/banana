package types

type BananaFile struct {
	TypeMeta `json:",inline" yaml:",inline"`
	MetaData *ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Name is the name of this konfig
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Version is the version of this konfig
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// Clusters is a list of clusters in this konf
	Clusters []*Cluster `json:"clusters,omitempty" yaml:"clusters,omitempty"`

	// Modules is a list of modules applied to this konfig
	Modules []Module `json:"modules,omitempty" yaml:"modules,omitempty"`
}
