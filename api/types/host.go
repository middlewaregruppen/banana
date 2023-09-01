package types

type Host struct {
	// Prefix is the prefix of this Host instance
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty"`

	// Wildcard is the wildcard of this Host instance
	Wildcard string `json:"wildcard,omitempty" yaml:"wildcard,omitempty"`

	// HostName is the hostname of this Host instance
	HostName string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	// Delimiter is the delimiter used to concatenate prefix, wildcard and hostname together
	Delimiter string `json:"delimiter,omitempty" yaml:"delimiter,omitempty"`
}
