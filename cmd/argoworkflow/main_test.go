package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmptyThen(t *testing.T) {
	tests := []struct {
		name   string
		first  string
		second string
		expect string
	}{{
		name:   "first is empty",
		first:  "",
		second: "second",
		expect: "second",
	}, {
		name:   "first is blank",
		first:  " ",
		second: "second",
		expect: " ",
	}}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EmptyThen(tt.first, tt.second)
			assert.Equal(t, tt.expect, result, "not match with", i)
		})
	}
}
