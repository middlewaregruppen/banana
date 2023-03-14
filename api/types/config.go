package types

type GlobalConfig struct {
	TypeMeta `json:",inline" yaml:",inline"`
	MetaData *ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}
