package resource

import (
	"fmt"
	"go/types"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestResourcesAreInitialised(t *testing.T) {
	assert.NotEmpty(t, Resources())

	loaded, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
	}, "github.com/aelsabbahy/goss/resource")
	require.NoError(t, err)

	found := map[string]types.TypeAndValue{}
	for _, pkg := range loaded {
		for t, typ := range pkg.TypesInfo.Types {
			if !typ.IsType() || typ.IsNil() || typ.IsBuiltin() {
				continue
			}
			if !strings.HasPrefix(typ.Type.String(), "github.com/aelsabbahy/goss/resource.") {
				continue
			}
			if !strings.HasPrefix(typ.Type.Underlying().String(), "struct") {
				continue
			}
			found[strings.ToLower(fmt.Sprintf("%s", t))] = typ
		}
	}
	foundAsArray := []string{}
	for name := range found {
		foundAsArray = append(foundAsArray, name)
	}
	actualAsArray := []string{}
	for name := range Resources() {
		actualAsArray = append(actualAsArray, name)
	}

	// subset, not equal, because 'found' contains some extraneous matches that are harder to filter out.
	assert.Subset(t, foundAsArray, actualAsArray)
}
