package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSay(t *testing.T) {
	value := say()

	assert.Equal(t, "Hello", value)
}
