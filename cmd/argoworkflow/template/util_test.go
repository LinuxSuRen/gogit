package template_test

import (
	"strings"
	"testing"
	"time"

	"github.com/linuxsuren/gogit/argoworkflow/template"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		object       interface{}
		expectErr    bool
		expectResult string
	}{{
		name:         "normal",
		template:     `{{.}}`,
		object:       "hello-template",
		expectErr:    false,
		expectResult: "hello-template",
	}, {
		name:         "template syntax error",
		template:     `{{.`,
		object:       "",
		expectErr:    true,
		expectResult: "",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := template.RenderTemplate(tt.template, tt.object)
			if tt.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestDuration(t *testing.T) {
	now := v1.Now()

	tests := []struct {
		name       string
		finishedAt v1.Time
		startedAt  v1.Time
		verify     func(t *testing.T, result string)
	}{{
		name:       "normal",
		finishedAt: now,
		startedAt:  v1.Time{Time: now.Add(time.Second * -3)},
		verify: func(t *testing.T, result string) {
			assert.Equal(t, "3s", result)
		},
	}, {
		name: "have not finished",
		startedAt: v1.Time{
			Time: now.Add(time.Millisecond * -1),
		},
		verify: func(t *testing.T, result string) {
			assert.True(t, strings.HasSuffix(result, "ms"), result)
		},
	}}
	for _, tt := range tests {
		result := template.Duration(tt.finishedAt, tt.startedAt)
		tt.verify(t, result)
	}
}
