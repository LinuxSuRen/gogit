package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	fakeruntime "github.com/linuxsuren/go-fake-runtime"
	"github.com/linuxsuren/gogit/pkg/oneapi"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

type commitOption struct {
	provider string
	token    string
	flags    []string
	runtime  fakeruntime.Execer
}

func newCommitCmd() *cobra.Command {
	opt := &commitOption{
		runtime: fakeruntime.NewDefaultExecer(),
	}
	cmd := &cobra.Command{
		Use:     "commit",
		RunE:    opt.runE,
		PreRunE: opt.preRunE,
		Short:   "Commit the current changes with AI",
		Long: `Commit the current changes with AI.
The AI provider is defined by the environment variable AI_PROVIDER,
and the token is defined by the environment variable ONEAPI_TOKEN.`,
	}
	cmd.Flags().StringSliceVarP(&opt.flags, "flag", "", []string{}, "The flags of the git commit command")
	return cmd
}

func (o *commitOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	o.provider = os.Getenv("AI_PROVIDER")
	if o.provider == "" {
		err = fmt.Errorf("AI_PROVIDER is not set")
	}
	o.token = os.Getenv("ONEAPI_TOKEN")
	if o.token == "" {
		err = errors.Join(err, fmt.Errorf("ONEAPI_TOKEN is not set"))
	}
	return
}

func (o *commitOption) runE(cmd *cobra.Command, args []string) (err error) {
	var gitdiff string
	gitdiff, err = o.getGitDiff()

	payload := oneapi.NewChatPayload(fmt.Sprintf("Please write a conventional git commit message for the following git diff:\n%s", gitdiff), "chatglm_std")

	var body []byte
	if body, err = json.Marshal(payload); err != nil {
		return
	}

	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1/chat/completions", o.provider), io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.token))
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusOK {
		// read the body and parse to oenapi.ChatResponse
		var chatResp oneapi.ChatResponse
		if err = json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
			return
		}

		var tempF *os.File
		if tempF, err = os.CreateTemp(os.TempDir(), "msg"); err != nil {
			return
		}

		content := chatResp.Choices[0].Message.Content
		// convert \n to new line
		content = strings.ReplaceAll(content, "\\n", "\n")
		if _, err = io.WriteString(tempF, content); err != nil {
			return
		}
		if err = tempF.Close(); err != nil {
			return
		}

		cmd.Println("start to commit with", tempF.Name())

		var gitExe string
		if gitExe, err = o.runtime.LookPath("git"); err != nil {
			return
		}

		if err = o.runtime.RunCommand(gitExe, "add", "."); err != nil {
			return
		}

		opts := []string{"git", "commit", "--edit", "--file", tempF.Name()}
		for _, flag := range o.flags {
			opts = append(opts, flag)
		}

		err = syscall.Exec(gitExe, opts, append(os.Environ(), "GIT_EDITOR=vim"))
	}
	return
}

func (o *commitOption) getGitDiff() (diff string, err error) {
	// run command git diff and get the output
	diff, err = o.runtime.RunCommandAndReturn("git", ".", "diff")
	return
}
