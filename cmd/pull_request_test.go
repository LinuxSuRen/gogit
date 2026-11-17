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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
		assert.Contains(t, strings.ToLower(err.Error()), "unsupported")
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

func TestSendToDingDing(t *testing.T) {
	originAPI := dingdingAPI
	defer func() {
		dingdingAPI = originAPI
	}()

	t.Run("message with special characters should be sent as valid JSON", func(t *testing.T) {
		var receivedBody []byte
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedBody, _ = io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"errcode": 0, "errmsg": "ok"}`))
		}))
		defer server.Close()
		dingdingAPI = server.URL

		msg := `msg with "quotes", \backslash and
newline`
		err := sendToDingDing("login", "token", msg, &scm.PullRequest{}, io.Discard)
		assert.Nil(t, err)

		assert.True(t, json.Valid(receivedBody), "the payload should be valid JSON")
		payload := map[string]interface{}{}
		assert.Nil(t, json.Unmarshal(receivedBody, &payload))
		text := payload["text"].(map[string]interface{})
		assert.Equal(t, msg, text["content"])
		// the token should not appear in the payload
		assert.NotContains(t, string(receivedBody), "access_token")
	})

	t.Run("error response code from dingding", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"errcode": 310000, "errmsg": "keyword not in content"}`))
		}))
		defer server.Close()
		dingdingAPI = server.URL

		err := sendToDingDing("login", "token", "hello", &scm.PullRequest{}, io.Discard)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "keyword not in content")
	})

	t.Run("unexpected http status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()
		dingdingAPI = server.URL

		err := sendToDingDing("login", "token", "hello", &scm.PullRequest{}, io.Discard)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "502")
	})
}
