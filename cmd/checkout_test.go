package cmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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

func TestFindRemoteURL(t *testing.T) {
	remotes := []*git.Remote{
		git.NewRemote(nil, &config.RemoteConfig{Name: "origin", URLs: []string{"https://example.com/origin.git"}}),
		git.NewRemote(nil, &config.RemoteConfig{Name: "upstream", URLs: []string{"https://example.com/upstream.git"}}),
	}

	assert.Equal(t, "https://example.com/origin.git", findRemoteURL(remotes, "origin"))
	assert.Equal(t, "https://example.com/upstream.git", findRemoteURL(remotes, "upstream"))
	// fall back to the first remote which has a URL
	assert.Equal(t, "https://example.com/origin.git", findRemoteURL(remotes, "not-exist"))

	// remotes without any URL should not panic
	emptyRemotes := []*git.Remote{
		git.NewRemote(nil, &config.RemoteConfig{Name: "origin"}),
	}
	assert.Empty(t, findRemoteURL(emptyRemotes, "origin"))
	assert.Empty(t, findRemoteURL(nil, "origin"))
}

// TestCheckoutWithoutRemote makes sure checkout fails with a clear error
// instead of panicking when the repository has no remote.
func TestCheckoutWithoutRemote(t *testing.T) {
	dir := t.TempDir()
	_, err := git.PlainInit(dir, false)
	assert.Nil(t, err)

	c := newCheckoutCommand()
	c.SetOut(io.Discard)
	c.SetErr(io.Discard)
	c.SetArgs([]string{"--target", dir})
	err = c.Execute()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cannot find any git remote")
}

// TestCheckoutDefaultBranch makes sure checkout works on repositories
// whose default branch is not master.
func TestCheckoutDefaultBranch(t *testing.T) {
	srcDir := t.TempDir()
	repo, err := git.PlainInit(srcDir, false)
	assert.Nil(t, err)

	wt, err := repo.Worktree()
	assert.Nil(t, err)
	assert.Nil(t, os.WriteFile(filepath.Join(srcDir, "README.md"), []byte("hello"), 0644))
	_, err = wt.Add("README.md")
	assert.Nil(t, err)
	sig := &object.Signature{Name: "gogit", Email: "gogit@example.com"}
	_, err = wt.Commit("init", &git.CommitOptions{Author: sig})
	assert.Nil(t, err)

	// rename the default branch to main
	head, err := repo.Head()
	assert.Nil(t, err)
	assert.Nil(t, repo.Storer.SetReference(
		plumbing.NewReferenceFromStrings("refs/heads/main", head.Hash().String())))
	assert.Nil(t, repo.Storer.SetReference(
		plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName("main"))))
	assert.Nil(t, repo.Storer.RemoveReference(plumbing.NewBranchReferenceName("master")))

	target := t.TempDir()
	versionFile := filepath.Join(t.TempDir(), "version")
	c := newCheckoutCommand()
	c.SetOut(io.Discard)
	c.SetErr(io.Discard)
	c.SetArgs([]string{srcDir, "--target", target, "--version-output", versionFile})
	err = c.Execute()
	assert.Nil(t, err)

	// the default branch should be checked out
	cloned, err := git.PlainOpen(target)
	assert.Nil(t, err)
	head, err = cloned.Head()
	assert.Nil(t, err)
	assert.Equal(t, "refs/heads/main", head.Name().String())

	// the version falls back to the HEAD commit instead of being empty
	version, err := os.ReadFile(versionFile)
	assert.Nil(t, err)
	assert.Equal(t, 8, len(version))
}
