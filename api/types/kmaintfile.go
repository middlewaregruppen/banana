package types

type KmaintFile struct {
	TypeMeta `json:",inline" yaml:",inline"`
	MetaData *ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Name is the name of this konfig
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Modules is a list of modules applied to this konfig
	Modules []Module `json:"modules,omitempty" yaml:"modules,omitempty"`
}

type Module struct {
	// Name is the name of this module
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Version is the version of this module
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// Opt is options that can be passed to this module
	Opts ModuleOpts `json:"opts,omitempty" yaml:"opts,omitempty"`
}

type ModuleOpts interface{}
