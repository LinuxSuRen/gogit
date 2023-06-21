package main

import (
	"fmt"
	"strings"

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
					Kind:     template.FileOutput,
					FileName: artifact.Path,
					File:     fmt.Sprintf("/artifact-files/%s/workflows/%s/%s/outputs/%s", wf.Namespace, wf.Name, node.ID, artifact.Name),
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

// GetOutputsWithTarget returns the outputs which has the target server address
func GetOutputsWithTarget(wf *wfv1.Workflow, target string) map[string]template.OutputObject {
	outputs := GetOutputs(wf)
	// add the server address
	if len(outputs) > 0 {
		for i, output := range outputs {
			if output.File != "" {
				output.File = target + output.File
			} else if strings.HasSuffix(i, ".md") {
				output.Kind = "markdown"
			}
			outputs[i] = output
		}
	}
	return outputs
}
