package pkg

import (
	"context"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReconcile(t *testing.T) {
	type args struct {
		ctx      context.Context
		repoInfo RepoInformation
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{{
		name: "pr number is -1",
		args: args{
			repoInfo: RepoInformation{
				PrNumber: -1,
			},
		},
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Reconcile(tt.args.ctx, tt.args.repoInfo); (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewStatusMaker(t *testing.T) {
	maker := NewStatusMaker("", "")
	assert.NotNil(t, maker)
	assert.NotNil(t, maker.expirationCheck)
	assert.False(t, maker.expirationCheck(nil, nil))
	assert.False(t, maker.expirationCheck(&scm.Status{State: scm.StateSuccess},
		&scm.StatusInput{State: scm.StateError}))
	assert.True(t, maker.expirationCheck(&scm.Status{State: scm.StateSuccess},
		&scm.StatusInput{State: scm.StateSuccess}))
}
