package template

// CommentTemplate is the default comment template
const CommentTemplate = `
[{{.Spec.WorkflowTemplateRef.Name}}]({{get .Annotations "workflow.templatelink"}}) is {{.Status.Phase}}. It started from {{date "01-02 15:04" .Status.StartedAt.Time}}, and took {{duration .Status.FinishedAt .Status.StartedAt}}. Please check log output from [here]({{get .Annotations "workflow.link"}}).

| Stage | Status | Duration |
|---|---|---|
{{- range $_, $status := .Status.Nodes}}
| {{$status.DisplayName}} | {{$status.Phase}} | {{duration $status.FinishedAt $status.StartedAt}} |
{{- end}}
`

const OutputsTemplate = `
Please feel free to check the following outputs:

| Output |
|---|
{{- range $name, $output := .}}
{{- if eq "file" (toString $output.Kind)}}
| [{{$output.FileName}}]({{$output.File}}) |
{{- else if eq "string" (toString $output.Kind)}}
| {{$name}}: {{$output.Value}} |
{{- end}}
{{- end}}

{{- range $name, $output := .}}
{{- if eq "markdown" (toString $output.Kind)}}
{{$output.Value}}
{{- end}}
{{- end}}
`
