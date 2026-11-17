package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"text/template"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/linuxsuren/gogit/pkg"
	"github.com/spf13/cobra"
)

func newPullRequestCmd() (c *cobra.Command) {
	opt := &pullRequestOption{}
	c = &cobra.Command{
		Use:     "pr",
		Short:   "Pull request related commands",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}
	opt.addFlags(c)
	flags := c.Flags()
	flags.BoolVarP(&opt.printAuthor, "author", "", false, "Print the author of the pull request")
	flags.BoolVarP(&opt.printReviewer, "reviewer", "", false, "Print the reviewers of the pull request")
	flags.BoolVarP(&opt.printAssignee, "assignee", "", false, "Print the assignees of the pull request")
	flags.StringVarP(&opt.msg, "msg", "", "", "The message of the pull request")
	flags.StringSliceVarP(&opt.dingdingTokenPairs, "dingding-tokens", "", []string{}, "The dingding token pairs of the pull request, format: login=token")
	flags.BoolVarP(&opt.skipInvalidPR, "skip-invalid-pr", "", true, "Skip the invalid pull request")
	flags.BoolVarP(&opt.strict, "strict", "", false, "Exit with error when any notification fails to be sent")
	return
}

func (o *pullRequestOption) preRunE(c *cobra.Command, args []string) (err error) {
	o.dingdingTokenMap = make(map[string]string)
	for _, pair := range o.dingdingTokenPairs {
		keyVal := strings.Split(pair, "=")
		if len(keyVal) == 2 {
			o.dingdingTokenMap[keyVal[0]] = keyVal[1]
		} else {
			err = fmt.Errorf("invalid dingding token pair: %q", pair)
			return
		}
	}
	return
}

func (o *pullRequestOption) runE(c *cobra.Command, args []string) (err error) {
	if o.pr <= 0 {
		if !o.skipInvalidPR {
			err = fmt.Errorf("invalid pr number %d", o.pr)
		}
		return
	}

	var scmClient *scm.Client
	if scmClient, err = o.getClient(); err != nil {
		return
	}

	var pr *scm.PullRequest
	if pr, _, err = scmClient.PullRequests.Find(c.Context(), o.repo, o.pr); err != nil {
		return
	}

	users := make(map[string]string, 0)
	addToMap(users, pr.Author.Login)

	if o.printAuthor {
		c.Println(pr.Author.Email, pr.Author.Name, pr.Author.Login, pr.Author.ID)
	}
	for _, user := range pr.Reviewers {
		if o.printReviewer {
			c.Println(user.Email, user.Name, user.Login, user.ID)
		}
		addToMap(users, user.Login)
	}
	for _, user := range pr.Assignees {
		if o.printAssignee {
			c.Println(user.Email, user.Name, user.Login, user.ID)
		}
		addToMap(users, user.Login)
	}

	var wait sync.WaitGroup
	var mu sync.Mutex
	var sendErrs []error
	for login := range users {
		wait.Add(1)
		go func(login string) {
			defer wait.Done()
			if token, ok := o.dingdingTokenMap[login]; ok {
				if sendErr := sendToDingDing(login, token, o.msg, pr, c.OutOrStderr()); sendErr != nil {
					mu.Lock()
					sendErrs = append(sendErrs, fmt.Errorf("failed to send the notification to %q: %v", login, sendErr))
					mu.Unlock()
				}
			}
		}(login)
	}
	wait.Wait()

	if len(sendErrs) > 0 {
		// notifications are best-effort by default, users can opt into failing via --strict
		if o.strict {
			err = errors.Join(sendErrs...)
		} else {
			c.PrintErrln(errors.Join(sendErrs...))
		}
	}
	return
}

// dingdingAPI is the DingTalk robot API address, kept as a variable for testing
var dingdingAPI = "https://oapi.dingtalk.com/robot/send"

func sendToDingDing(login, token, msg string, pr *scm.PullRequest, c io.Writer) (err error) {
	formattedMsg, fmtErr := formatMessage(msg, pr)
	if fmtErr != nil {
		log.Printf("cannot format the message %q: %v\n", msg, fmtErr)
		formattedMsg = msg
	}

	api := fmt.Sprintf("%s?access_token=%s", dingdingAPI, token)
	payload, mErr := json.Marshal(map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": formattedMsg},
	})
	if mErr != nil {
		return fmt.Errorf("cannot marshal the notification payload: %v", mErr)
	}

	var resp *http.Response
	if resp, err = http.Post(api, "application/json", bytes.NewReader(payload)); err != nil {
		_, _ = fmt.Fprintln(c, err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		_, _ = fmt.Fprintf(c, "send message to %q failed, received code %d instead of 200\n", login, resp.StatusCode)
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	dingdingResp := &DingDingResponse{}
	if uErr := json.Unmarshal(body, dingdingResp); uErr != nil {
		_, _ = fmt.Fprintf(c, "cannot unmarshal the response for %q: %v\n", login, uErr)
		return uErr
	} else if dingdingResp.ErrCode != 0 {
		_, _ = fmt.Fprintf(c, "receive error response for %q: %q\n", login, dingdingResp.ErrMsg)
		return fmt.Errorf("receive error response: %q", dingdingResp.ErrMsg)
	}

	_, _ = fmt.Fprintf(c, "send message to %q successfully\n", login)
	return
}

func formatMessage(msg string, pr *scm.PullRequest) (result string, err error) {
	var tpl *template.Template
	if tpl, err = template.New("message").Parse(msg); err == nil {
		var b strings.Builder
		if err = tpl.Execute(&b, pr); err == nil {
			result = b.String()
		}
		err = pkg.WrapError(err, "cannot format the message %q: %v", msg)
		return
	}
	return
}

type DingDingResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func addToMap(m map[string]string, k string) {
	if _, ok := m[k]; !ok {
		m[k] = ""
	}
}

type pullRequestOption struct {
	gitProviderOption
	printAuthor        bool
	printReviewer      bool
	printAssignee      bool
	msg                string
	dingdingTokenPairs []string
	dingdingTokenMap   map[string]string
	skipInvalidPR      bool
	strict             bool
}
