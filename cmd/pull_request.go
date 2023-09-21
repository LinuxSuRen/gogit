package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/jenkins-x/go-scm/scm"
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
				api := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", token)
				msg := strings.NewReader(`{"msgtype": "text", "text": {"content": "` + o.msg + `"}}`)

				resp, err := http.Post(api, "application/json", msg)
				if err != nil {
					c.Println(err)
				} else if resp.StatusCode != http.StatusOK {
					c.Printf("send message to %q failed, received code %d instead of 200\n", login, resp.StatusCode)
				} else {
					body, _ := io.ReadAll(resp.Body)

					dingdingResp := &DingDingResponse{}
					if err := json.Unmarshal(body, dingdingResp); err != nil {
						c.Printf("cannot unmarshal the response for %q: %v\n", login, err)
					} else if dingdingResp.ErrCode != 0 {
						c.Printf("receive error response for %q: %q\n", login, dingdingResp.ErrMsg)
					} else {
						c.Printf("send message to %q successfully\n", login)
					}

				}
			}
		}(login)
	}
	wait.Wait()
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
}
