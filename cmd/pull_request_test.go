package cmd

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPullRequestCmd(t *testing.T) {
	t.Run("missing required flags", func(t *testing.T) {
		c := NewRootCommand()
		c.SetOut(io.Discard)

		c.SetArgs([]string{"pr", "--dingding-tokens", "a=b"})
		err := c.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required flag(s)")
	})

	t.Run("invalid dingding token pairs", func(t *testing.T) {
		c := NewRootCommand()
		c.SetOut(io.Discard)

		c.SetArgs([]string{"pr", "--dingding-tokens", "a=b=c"})
		err := c.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid dingding token pair")
	})

	t.Run("invalid git kind", func(t *testing.T) {
		c := NewRootCommand()
		c.SetOut(io.Discard)

		c.SetArgs([]string{"pr", "--provider", "invalid", "--pr=1", "--repo=xxx/xxx", "--token=token", "--username=xxx"})
		err := c.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Unsupported")
	})
}
