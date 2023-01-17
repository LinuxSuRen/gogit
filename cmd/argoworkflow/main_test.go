package main

import (
	"testing"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/stretchr/testify/assert"
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

func TestDeleteRetryNodes(t *testing.T) {
	wf := &wfv1.Workflow{
		Status: wfv1.WorkflowStatus{
			Nodes: map[string]wfv1.NodeStatus{},
		},
	}
	wf.Status.Nodes["test"] = wfv1.NodeStatus{DisplayName: "test"}
	wf.Status.Nodes["test-0"] = wfv1.NodeStatus{DisplayName: "test-(0)"}
	wf.Status.Nodes["test-10"] = wfv1.NodeStatus{DisplayName: "test-(10)"}
	wf.Status.Nodes["test-100"] = wfv1.NodeStatus{DisplayName: "test-(100)"}

	deleteRetryNodes(wf)
	assert.Equal(t, 1, len(wf.Status.Nodes))
}
