package types

type KmaintFile struct {
	TypeMeta `json:",inline" yaml:",inline"`
	MetaData *ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Name is the name of this konfig
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Modules is a list of modules applied to this konfig
	Modules []Module `json:"modules,omitempty" yaml:"modules,omitempty"`
}
