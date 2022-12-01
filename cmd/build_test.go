package cmd

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestStatus(t *testing.T) {
	tests := []struct {
		name         string
		opt          *option
		status       string
		expectStatus string
	}{{
		name:         "Capital letter",
		status:       "Running",
		expectStatus: "running",
	}, {
		name:         "Special token: Succeeded",
		status:       "Succeeded",
		expectStatus: "success",
	}}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			opt := &option{status: tt.status}
			_ = opt.preRunE(nil, nil)
			assert.Equal(t, tt.expectStatus, opt.status)
		})
	}
}

func TestTokenFromFile(t *testing.T) {
	tests := []struct {
		name    string
		opt     *option
		token   string
		prepare func() string
		expect  string
	}{{
		name:   "plain token",
		token:  "token",
		expect: "token",
	}, {
		name: "token from file",
		prepare: func() string {
			f, _ := os.CreateTemp(os.TempDir(), "token")
			_, _ = io.WriteString(f, "token")
			return f.Name()
		},
		expect: "token",
	}}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			opt := &option{token: tt.token}
			if tt.prepare != nil {
				token := tt.prepare()
				opt.token = "file://" + token
				defer func() {
					_ = os.RemoveAll(token)
				}()
			}
			_ = opt.preRunE(nil, nil)
			assert.Equal(t, tt.expect, opt.token)
		})
	}
}

func TestFlags(t *testing.T) {
	cmd := NewBuildCmd()
	assert.Equal(t, "gogit", cmd.Use)
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("provider"))
	assert.NotNil(t, flags.Lookup("server"))
	assert.NotNil(t, flags.Lookup("owner"))
	assert.NotNil(t, flags.Lookup("repo"))
	assert.NotNil(t, flags.Lookup("pr"))
	assert.NotNil(t, flags.Lookup("username"))
	assert.NotNil(t, flags.Lookup("token"))
	assert.NotNil(t, flags.Lookup("status"))
	assert.NotNil(t, flags.Lookup("label"))
	assert.NotNil(t, flags.Lookup("description"))
}
