package adx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtils_unescapeEntityName(t *testing.T) {
	name := unescapeEntityName("name")
	assert.Equal(t, "name", name, "an already-unescaped name should have been left alone")

	name = unescapeEntityName("['name']")
	assert.Equal(t, "name", name, "name should have had adx escaping removed")
}
