package main

import (
	"fmt"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/linuxsuren/gogit/argoworkflow/template"
)

// GetOutputs returns the outputs in map format.
// Key is the node name, value is the output link.
// The output link address does not have the host part.
func GetOutputs(wf *wfv1.Workflow) (outputs map[string]template.OutputObject) {
	outputs = map[string]template.OutputObject{}
	nodes := wf.Status.Nodes
	for _, node := range nodes {
		if node.Outputs == nil {
			continue
		}

		var outputObject *template.OutputObject
		for _, artifact := range node.Outputs.Artifacts {
			key := fmt.Sprintf("%s-%s", node.Name, artifact.Name)

			if artifact.Path != "" {
				// TODO assume this is a artifact file
				outputObject = &template.OutputObject{
					Kind: template.FileOutput,
					File: fmt.Sprintf("/artifact-files/%s/workflows/%s/%s/outputs/%s", wf.Namespace, wf.Name, node.Name, artifact.Name),
				}
				outputs[key] = *outputObject
			}
		}

		for _, param := range node.Outputs.Parameters {
			key := fmt.Sprintf("%s-%s", node.Name, param.Name)

			if param.Value != nil && param.Value.String() != "" {
				outputObject = &template.OutputObject{
					Kind:  template.ValueOutput,
					Value: param.Value.String(),
				}
				outputs[key] = *outputObject
			}
		}
	}
	return
}
