package main

import (
	"strings"
	"testing"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/linuxsuren/gogit/argoworkflow/template"
	"github.com/stretchr/testify/assert"
)

func TestGetOutputsWithTarget(t *testing.T) {
	emptyWf := &wfv1.Workflow{
		Status: wfv1.WorkflowStatus{
			Nodes: wfv1.Nodes{
				"test": wfv1.NodeStatus{},
			},
		},
	}
	assert.Equal(t, 0, len(GetOutputsWithTarget(emptyWf, "")))

	wf := &wfv1.Workflow{
		Status: wfv1.WorkflowStatus{
			Nodes: wfv1.Nodes{
				"test": wfv1.NodeStatus{
					Name: "test",
					Outputs: &wfv1.Outputs{
						Artifacts: wfv1.Artifacts{{
							Path: "test/install.yaml",
							Name: "test",
						}},
					},
				},
				"report_md": wfv1.NodeStatus{
					Name: "report_md",
					Outputs: &wfv1.Outputs{
						Parameters: []wfv1.Parameter{{
							Name:  "report_md",
							Value: wfv1.AnyStringPtr("## Report\n"),
						}},
					},
				},
			},
		},
	}

	outputs := GetOutputsWithTarget(wf, "https://github.com")
	for _, output := range outputs {
		if output.Kind == "file" {
			assert.Truef(t, strings.HasPrefix(output.File, "https://github.com"), output.File)
		} else if output.Kind == "markdown" {
			assert.Equal(t, "## Report\n", output.Value)
		}
	}
	assert.Equal(t, map[string]template.OutputObject{
		"test-test": {
			Kind:     template.FileOutput,
			File:     "https://github.com/artifact-files//workflows///outputs/test",
			FileName: "test/install.yaml",
		},
		"report_md-report_md": {
			Kind:  template.MarkdownOutput,
			Value: "## Report\n",
		},
	}, outputs)
}
