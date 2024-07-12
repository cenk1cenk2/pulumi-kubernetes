package kinds

import (
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// IsPatchURN returns true if the URN is for a Patch resource.
func IsPatchURN(urn resource.URN, kind string) bool {
	urnS := urn.Type().String()

	// Existing logic based on hardcoded lookup.
	if PatchQualifiedTypes.Has(urnS) {
		return true
	}

	// New logic based on kind.
	return strings.HasSuffix(urnS, "Patch") && !strings.HasSuffix(kind, "Patch")
}

// IsListURN returns true if the URN is for a List resource.
func IsListURN(urn resource.URN) bool {
	urnS := urn.Type().String()

	return ListQualifiedTypes.Has(urnS)
}
