package cmd

import (
	"github.com/stretchr/testify/assert"
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
		name:         "Special status: Succeeded",
		status:       "Succeeded",
		expectStatus: "success",
	}}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			opt := &option{status: tt.status}
			opt.preRun(nil, nil)
			assert.Equal(t, tt.expectStatus, opt.status)
		})
	}
}
