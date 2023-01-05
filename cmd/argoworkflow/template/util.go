package template

import (
	"bytes"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RenderTemplate renders the template, and returns the result text
func RenderTemplate(text string, object interface{}) (result string, err error) {
	var tpl *template.Template
	if tpl, err = template.New("temp").Funcs(sprig.TxtFuncMap()).Funcs(map[string]any{
		"durationStr": durationStr,
		"duration":    Duration,
		"get":         getValue,
	}).Parse(text); err != nil {
		return
	}

	data := new(bytes.Buffer)
	if err = tpl.Execute(data, object); err == nil {
		result = data.String()
	}
	return
}

func getValue(data map[string]string, key string) string {
	return data[key]
}

func durationStr(finishedAt, startedAt string) string {
	finishedTime, _ := time.Parse(time.RFC3339, finishedAt)
	startedTime, _ := time.Parse(time.RFC3339, startedAt)
	return Duration(v1.Time{Time: finishedTime}, v1.Time{Time: startedTime})
}

// Duration returns the duration time string
func Duration(finishedAt, startedAt v1.Time) string {
	if finishedAt.Before(&startedAt) {
		// have not finished
		finishedAt = v1.Now()
	}
	duration := finishedAt.Sub(startedAt.Time)
	return duration.String()
}
