/*
MIT License

Copyright (c) 2023 Rick

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package cmd

import (
	"io"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
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

	t.Run("invalid pr number, do not skip", func(t *testing.T) {
		c := NewRootCommand()
		c.SetOut(io.Discard)

		c.SetArgs([]string{"pr", "--pr=-1", "--repo=xxx/xxx", "--token=token", "--username=xxx", "--skip-invalid-pr=false"})
		err := c.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pr number")
	})

	t.Run("skip invalid pr number", func(t *testing.T) {
		c := NewRootCommand()
		c.SetOut(io.Discard)

		c.SetArgs([]string{"pr", "--pr=-1", "--repo=xxx/xxx", "--token=token", "--username=xxx"})
		err := c.Execute()
		assert.NoError(t, err)
	})
}

func TestFormatMessage(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		pr     *scm.PullRequest
		expect string
	}{{
		name: "normal",
		msg:  "hello {{.Title}}",
		pr: &scm.PullRequest{
			Title: "world",
		},
		expect: "hello world",
	}, {
		name:   "invalid template syntax",
		msg:    "hello {{.Title}",
		expect: "",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := formatMessage(tt.msg, tt.pr)
			assert.Equal(t, tt.expect, result)
		})
	}
}
