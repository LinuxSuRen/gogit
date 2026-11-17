package main

import (
	"context"
	"encoding/json"
	"testing"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/pkg/plugins/executor"
	"github.com/linuxsuren/gogit/pkg"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestOptionString(t *testing.T) {
	opt := option{
		Provider: "github",
		Token:    "super-secret-token",
		Target:   "http://argo",
	}

	text := opt.String()
	assert.Contains(t, text, "github")
	assert.NotContains(t, text, "super-secret-token")
	assert.Contains(t, text, "su***en")
}

func TestMaskString(t *testing.T) {
	assert.Equal(t, "***", maskString(""))
	assert.Equal(t, "***", maskString("abc"))
	assert.Equal(t, "ab***yz", maskString("abcdefghijklmnopqrstuvwxyz"))
}

func newTestWorkflow() (wf *wfv1.Workflow, args executor.ExecuteTemplateArgs) {
	wf = &wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "wf",
		},
		Status: wfv1.WorkflowStatus{
			Phase: wfv1.WorkflowSucceeded,
			Nodes: map[string]wfv1.NodeStatus{
				"wf": {Phase: wfv1.NodeSucceeded},
			},
		},
	}

	args = executor.ExecuteTemplateArgs{
		Workflow: &executor.Workflow{
			ObjectMeta: executor.ObjectMeta{
				Namespace: "default",
				Name:      "wf",
			},
		},
		Template: &wfv1.Template{
			Plugin: &wfv1.Plugin{},
		},
	}
	return
}

func TestExecuteWithInvalidPR(t *testing.T) {
	wf, args := newTestWorkflow()
	args.Template.Plugin.Value = json.RawMessage(`{"gogit-executor-plugin": {"pr": "3a"}}`)

	executorPlugin := &DefaultPluginExecutor{option: &option{}}
	resp, err := executorPlugin.Execute(args, wf)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), `wrong pull-request number "3a"`)
	assert.Equal(t, wfv1.NodeFailed, resp.Body.Node.Phase)
}

func TestHandleCommentWithoutWorkflowTemplateRef(t *testing.T) {
	wf, args := newTestWorkflow()

	opt := &pluginOption{Option: &option{Target: "http://argo"}}
	repo := pkg.RepoInformation{
		// -1 means do not send the comment to a real git provider
		PrNumber: -1,
	}

	err := handleComment(context.Background(), args, wf, opt, repo, "http://argo/workflows/default/wf")
	assert.Nil(t, err)
	assert.Equal(t, "http://argo/workflows/default/wf", wf.Annotations["workflow.link"])
	assert.Empty(t, wf.Annotations["workflow.templatelink"])
}
