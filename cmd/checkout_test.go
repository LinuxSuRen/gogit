package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_prRef(t *testing.T) {
	type args struct {
		pr   int
		kind string
	}
	tests := []struct {
		name    string
		args    args
		wantRef string
	}{{
		name: "gitlab",
		args: args{
			pr:   1,
			kind: "gitlab",
		},
		wantRef: "refs/merge-requests/1/head:pr-1",
	}, {
		name: "unknown",
		args: args{
			pr:   1,
			kind: "unknown",
		},
		wantRef: "",
	}, {
		name: "github",
		args: args{
			pr:   1,
			kind: "github",
		},
		wantRef: "refs/pull/1/head:pr-1",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantRef, prRef(tt.args.pr, tt.args.kind), "prRef(%v, %v)", tt.args.pr, tt.args.kind)
		})
	}
}

func Test_detectGitKind(t *testing.T) {
	type args struct {
		gitURL string
	}
	tests := []struct {
		name     string
		args     args
		wantKind string
	}{{
		name: "github",
		args: args{
			gitURL: "https://github.com/linuxsuren/gogit",
		},
		wantKind: "github",
	}, {
		name: "gitlab",
		args: args{
			gitURL: "git@10.121.218.82:demo/test.git",
		},
		wantKind: "gitlab",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantKind, detectGitKind(tt.args.gitURL), "detectGitKind(%v)", tt.args.gitURL)
		})
	}
}

func TestGetAuth(t *testing.T) {
	opt := &checkoutOption{
		sshPrivateKey: "/tmp",
	}
	auth, err := opt.getAuth("git@fake.com")
	assert.Nil(t, auth)
	assert.NotNil(t, err)

	auth, err = opt.getAuth("fake.com")
	assert.Nil(t, auth)
	assert.Nil(t, err)
}

func TestPreRunE(t *testing.T) {
	const sampleGit = "https://github.com/linuxsuren/gogit"
	const anotherGit = "https://github.com/linuxsuren/gogit.git"

	tests := []struct {
		name      string
		opt       *checkoutOption
		args      []string
		expectErr bool
		verify    func(t *testing.T, opt *checkoutOption)
	}{{
		name:      "url is empty",
		opt:       &checkoutOption{},
		args:      []string{sampleGit},
		expectErr: false,
		verify: func(t *testing.T, opt *checkoutOption) {
			assert.Equal(t, sampleGit, opt.url)
		},
	}, {
		name:      "url is not empty",
		opt:       &checkoutOption{url: anotherGit},
		args:      []string{sampleGit},
		expectErr: false,
		verify: func(t *testing.T, opt *checkoutOption) {
			assert.Equal(t, anotherGit, opt.url)
		},
	}, {
		name:      "branch is fullname",
		opt:       &checkoutOption{branch: "refs/heads/master"},
		expectErr: false,
		verify: func(t *testing.T, opt *checkoutOption) {
			assert.Equal(t, "master", opt.branch)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opt.preRunE(nil, tt.args)
			if tt.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			tt.verify(t, tt.opt)
		})
	}
}
