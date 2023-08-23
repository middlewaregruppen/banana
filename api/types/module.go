package types

type Module struct {
	MetaData *ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Name is the name of this module
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Version is the version of this module
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// Namespace is the namespace for this module
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// Opt is options that can be passed to this module
	Opts ModuleOpts `json:"opts,omitempty" yaml:"opts,omitempty"`

	// Components is a list of components for this module
	Components []Component `json:"components,omitempty" yaml:"components,omitempty"`
}

type ModuleOpts map[string]interface{}
