package types

type Age struct {
	// Recipients is a list of age recipients
	Recipients []string `json:"recipients,omitempty" yaml:"recipients,omitempty"`
}
