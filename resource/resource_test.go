package resource

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourcesAreInitialised(t *testing.T) {
	assert.NotEmpty(t, Resources())

	var i Resource
	it := reflect.TypeOf(i)
	fmt.Printf("it: %v", it)
}
