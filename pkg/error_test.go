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

package pkg_test

import (
	"errors"
	"testing"

	"github.com/linuxsuren/gogit/pkg"
	"github.com/stretchr/testify/assert"
)

func TestWrapError(t *testing.T) {
	tests := []struct {
		err    error
		msg    string
		args   []any
		expect error
	}{{
		err:    nil,
		expect: nil,
	}, {
		err:    defaultErr,
		msg:    "wrap %v",
		expect: errors.New("wrap fake"),
	}}
	for _, tt := range tests {
		err := pkg.WrapError(tt.err, tt.msg, tt.args...)
		assert.Equal(t, tt.expect, err)
	}
}

func TestIgnoreError(t *testing.T) {
	assert.Equal(t, nil, pkg.IgnoreError(nil, "fake"))
	assert.Equal(t, nil, pkg.IgnoreError(defaultErr, "fake"))
	assert.Equal(t, defaultErr, pkg.IgnoreError(defaultErr, "other"))
}

var defaultErr = errors.New("fake")
