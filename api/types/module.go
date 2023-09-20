package types

type Module struct {
	MetaData *ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Name is the name of this module
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Version is the version of this module, typically translates to a git tag
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// Ref is the git reference name of this module
	Ref string `json:"ref,omitempty" yaml:"ref,omitempty"`

	// Namespace is the namespace for this module
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// Opt is options that can be passed to this module
	Opts ModuleOpts `json:"opts,omitempty" yaml:"opts,omitempty"`

	// Components is a list of components for this module
	Components []string `json:"components,omitempty" yaml:"components,omitempty"`

	// Hosts is a list of Host types mapped to this module
	Host *Host `json:"hosts,omitempty" yaml:"hosts,omitempty"`

	// Secrets is a list of secrets mapped to this module
	Secrets []string `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

type ModuleOpts map[string]interface{}
