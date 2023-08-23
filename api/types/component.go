package types

type Component struct {
	// Name is the name of this component
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Version is the version of this component
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}
