package module

type Secret struct {
	Key   string
	Value string
}

// IsFile returns true if secret is a path to a file. This is determined
// by the key. If it is prefixed with '@' then the secret is assumed to be a path.
func (s *Secret) IsFile() bool {
	return s.Key[0:1] == "@"
}
