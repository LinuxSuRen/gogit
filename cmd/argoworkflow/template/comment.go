package template

// CommentTemplate is the default comment template
const CommentTemplate = `
{{.Spec.WorkflowTemplateRef.Name}} is {{.Status.Phase}}. It takes {{duration .Status.FinishedAt .Status.StartedAt}}. Please check log output from [here]({{get .Annotations "workflow.link"}}).

| Stage | Status | Duration |
|---|---|---|
{{- range $_, $status := .Status.Nodes}}
| {{$status.DisplayName}} | {{$status.Phase}} | {{duration $status.FinishedAt $status.StartedAt}} |
{{- end}}
`
