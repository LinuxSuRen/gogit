package pkg

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/stretchr/testify/assert"
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
			if err := NewMaker(tt.args.ctx, tt.args.repoInfo); (err != nil) != tt.wantErr {
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

func TestGetCommentIDs(t *testing.T) {
	tests := []struct {
		comments []*scm.Comment
		expect   []int
	}{{
		comments: []*scm.Comment{{
			Body: "start",
			ID:   1,
		}, {
			Body: "start - end",
			ID:   2,
		}, {
			Body: "other - end",
			ID:   3,
		}, {
			Body: "other",
			ID:   4,
		}, {
			Body: "good - end",
			ID:   5,
		}},
		expect: []int{2, 3, 5},
	}}
	for _, tt := range tests {
		assert.Equal(t, tt.expect, getCommentIDs(tt.comments, "end"))
	}
}
