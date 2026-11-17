package cmd

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	c := NewRootCommand()
	assert.NotNil(t, c)
	assert.Equal(t, "gogit", c.Use)
}

func TestVersionCmd(t *testing.T) {
	c := NewRootCommand()
	out := io.Discard
	c.SetOut(out)
	c.SetErr(out)
	c.SetArgs([]string{"version"})
	assert.Nil(t, c.Execute())

	versionCmd := c.Commands()
	found := false
	for _, sub := range versionCmd {
		if sub.Name() == "version" {
			found = true
		}
	}
	assert.True(t, found, "the version sub-command should be registered")
}
