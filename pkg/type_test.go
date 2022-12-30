package pkg

import "testing"

func TestRepoInformation_getRepoPath(t *testing.T) {
	type fields struct {
		Owner string
		Repo  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name: "normal case",
		fields: fields{
			Owner: "owner",
			Repo:  "repo",
		},
		want: "owner/repo",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RepoInformation{
				Owner: tt.fields.Owner,
				Repo:  tt.fields.Repo,
			}
			if got := r.getRepoPath(); got != tt.want {
				t.Errorf("getRepoPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
