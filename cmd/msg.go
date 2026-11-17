package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	fakeruntime "github.com/linuxsuren/go-fake-runtime"
	"github.com/linuxsuren/gogit/pkg/oneapi"
	"github.com/spf13/cobra"
)

// defaultAIModel is the default model of the AI provider
const defaultAIModel = "chatglm_std"

type commitOption struct {
	provider string
	token    string
	model    string
	edit     bool
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
and the token is defined by the environment variable ONEAPI_TOKEN.
The model is defined by the environment variable AI_MODEL.`,
	}
	cmd.Flags().StringVarP(&opt.model, "model", "m", "",
		fmt.Sprintf("The model of the AI provider, or via the environment variable AI_MODEL (default %q)", defaultAIModel))
	cmd.Flags().BoolVarP(&opt.edit, "edit", "e", false,
		"Open the editor to edit the commit message before committing")
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
	if o.model == "" {
		o.model = os.Getenv("AI_MODEL")
	}
	if o.model == "" {
		o.model = defaultAIModel
	}
	return
}

// httpClient has a timeout to avoid hanging forever on a slow AI provider
var httpClient = &http.Client{
	Timeout: time.Minute * 3,
}

func (o *commitOption) runE(cmd *cobra.Command, args []string) (err error) {
	var gitdiff string
	if gitdiff, err = o.getGitDiff(); err != nil {
		return
	}

	payload := oneapi.NewChatPayload(fmt.Sprintf("Please write a conventional git commit message for the following git diff:\n%s", gitdiff), o.model)

	var body []byte
	if body, err = json.Marshal(payload); err != nil {
		return
	}

	var req *http.Request
	if req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1/chat/completions", o.provider),
		io.NopCloser(bytes.NewReader(body))); err != nil {
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.token))
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	if resp, err = httpClient.Do(req); err != nil {
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var respBody []byte
	if respBody, err = io.ReadAll(resp.Body); err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the AI provider returned an unexpected status code %d, response: %s",
			resp.StatusCode, truncate(string(respBody), 200))
	}

	var chatResp oneapi.ChatResponse
	if err = json.Unmarshal(respBody, &chatResp); err != nil {
		return
	}
	if len(chatResp.Choices) == 0 {
		return fmt.Errorf("the AI provider returned no choice, please check the model %q and the token", o.model)
	}

	var tempF *os.File
	if tempF, err = os.CreateTemp(os.TempDir(), "msg"); err != nil {
		return
	}
	tempFile := tempF.Name()
	defer func() {
		_ = tempF.Close()
		_ = os.Remove(tempFile)
	}()

	content := chatResp.Choices[0].Message.Content
	// convert \n to new line
	content = strings.ReplaceAll(content, "\\n", "\n")
	if _, err = io.WriteString(tempF, content); err != nil {
		return
	}
	if err = tempF.Close(); err != nil {
		return
	}

	cmd.Println("start to commit with", tempFile)

	var gitExe string
	if gitExe, err = o.runtime.LookPath("git"); err != nil {
		return
	}

	if err = o.runtime.RunCommand(gitExe, "add", "."); err != nil {
		return
	}

	opts := []string{"commit", "--file", tempFile}
	if o.edit {
		opts = append(opts, "--edit")
	}
	opts = append(opts, o.flags...)

	// run git commit as a child process so that it works on all the supported platforms
	err = o.runtime.RunCommand(gitExe, opts...)
	return
}

func (o *commitOption) getGitDiff() (diff string, err error) {
	// run command git diff and get the output
	diff, err = o.runtime.RunCommandAndReturn("git", ".", "diff")
	return
}

// truncate cuts the text to the given size
func truncate(text string, size int) string {
	if len(text) <= size {
		return text
	}
	return text[:size] + "..."
}
