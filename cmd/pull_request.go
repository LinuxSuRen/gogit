package cmd

import (
	"encoding/json"
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
	for login, _ := range users {
		wait.Add(1)
		go func(login string) {
			defer wait.Done()
			if token, ok := o.dingdingTokenMap[login]; ok {
				sendToDingDing(login, token, o.msg, pr, c.OutOrStderr())
			}
		}(login)
	}
	wait.Wait()
	return
}

func sendToDingDing(login, token, msg string, pr *scm.PullRequest, c io.Writer) (err error) {
	formattedMsg, fmtErr := formatMessage(msg, pr)
	if fmtErr != nil {
		log.Printf("cannot format the message %q: %v\n", msg, fmtErr)
		formattedMsg = msg
	}

	api := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", token)
	payload := strings.NewReader(`{"msgtype": "text", "text": {"content": "` + formattedMsg + `"}}`)
	fmt.Println(`{"msgtype": "text", "text": {"content": "`+formattedMsg+`"}}`, api)

	var resp *http.Response
	resp, err = http.Post(api, "application/json", payload)
	if err != nil {
		fmt.Fprintln(c, err)
	} else if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(c, "send message to %q failed, received code %d instead of 200\n", login, resp.StatusCode)
	} else {
		body, _ := io.ReadAll(resp.Body)

		dingdingResp := &DingDingResponse{}
		if err := json.Unmarshal(body, dingdingResp); err != nil {
			fmt.Fprintf(c, "cannot unmarshal the response for %q: %v\n", login, err)
		} else if dingdingResp.ErrCode != 0 {
			fmt.Fprintf(c, "receive error response for %q: %q\n", login, dingdingResp.ErrMsg)
		} else {
			fmt.Fprintf(c, "send message to %q successfully\n", login)
		}
	}
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
}
