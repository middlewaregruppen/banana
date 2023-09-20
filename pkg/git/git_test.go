package git

import (
	"strings"
	"testing"
)

func parseRevision(ref string) string {
	refParts := strings.SplitN(ref, "/", 3)
	return refParts[len(refParts)-1]
}

func TestParseRevision(t *testing.T) {

}
