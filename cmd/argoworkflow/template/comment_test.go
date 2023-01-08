package template_test

import (
	"testing"
	"time"

	"github.com/linuxsuren/gogit/argoworkflow/template"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCommentTemlate(t *testing.T) {
	startTime := v1.Now()
	endTime := v1.Time{Time: startTime.Add(time.Second * 5)}

	node1 := map[string]interface{}{"Phase": "Success", "DisplayName": "node-1", "StartedAt": startTime, "FinishedAt": endTime, "StartedTime": "2023-01-06T07:49:07Z", "EndedTime": "2023-01-06T07:54:26Z"}
	node2 := map[string]interface{}{"Phase": "Failed", "DisplayName": "node-2", "StartedAt": startTime, "FinishedAt": endTime, "StartedTime": "2023-01-06T07:49:07Z", "EndedTime": "2023-01-06T07:54:26Z"}
	// TODO cannot handle the situation in template
	// node3 := map[string]interface{}{"DisplayName": "node-2.OnExit"}
	// node4 := map[string]interface{}{"DisplayName": "plugin-8bs4n.hooks.all"}

	nodes := map[string]interface{}{}
	nodes["node1"] = node1
	nodes["node2"] = node2
	// nodes["node3"] = node3
	// nodes["node4"] = node4

	status := map[string]interface{}{}
	status["Nodes"] = nodes
	status["Phase"] = "Failed"
	status["StartedAt"] = startTime
	status["FinishedAt"] = endTime

	object := map[string]interface{}{}
	object["Status"] = status
	object["Annotations"] = map[string]string{
		"workflow.link":         "https://github.com/linxusuren/gogit",
		"workflow.templatelink": "https://github.com/linxusuren/gogit.git",
	}
	object["Spec"] = map[string]interface{}{
		"WorkflowTemplateRef": map[string]string{
			"Name": "Sample",
		},
	}

	result, err := template.RenderTemplate(template.CommentTemplate, object)
	assert.Nil(t, err)
	assert.Equal(t, `
[Sample](https://github.com/linxusuren/gogit.git) is Failed. It takes 5s. Please check log output from [here](https://github.com/linxusuren/gogit).

| Stage | Status | Duration |
|---|---|---|
| node-1 | Success | 5s |
| node-2 | Failed | 5s |
`, result)

	result, err = template.RenderTemplate(`
| Stage | Status | Duration |
|---|---|---|
{{ range $node, $status := .Status.Nodes -}}
| {{$status.DisplayName}} | {{$status.Phase}} | {{durationStr $status.EndedTime $status.StartedTime}} |
{{end -}}
`, object)
	assert.Nil(t, err)
	assert.Equal(t, `
| Stage | Status | Duration |
|---|---|---|
| node-1 | Success | 5m19s |
| node-2 | Failed | 5m19s |
`, result)
}
