package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRootCommand(t *testing.T) {
	c := NewRootCommand()
	assert.NotNil(t, c)
	assert.Equal(t, "gogit", c.Use)
}
