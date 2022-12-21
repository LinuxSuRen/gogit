package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/pkg/plugins/executor"
	"github.com/linuxsuren/gogit/pkg"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"os"
	"strconv"
)

func main() {
	opt := &option{}
	cmd := &cobra.Command{
		Use:  "workflow-executor-gogit",
		RunE: opt.runE,
	}
	flags := cmd.Flags()
	flags.StringVarP(&opt.Provider, "provider", "", "",
		"Git provider, such as: gitlab/github")
	flags.StringVarP(&opt.Server, "server", "", "",
		"Git server address, only required when it's not a public service")
	flags.StringVarP(&opt.Username, "username", "", "",
		"Username of the git server")
	flags.StringVarP(&opt.Token, "token", "", "",
		"Personal access token of  the git server")
	flags.IntVarP(&opt.Port, "port", "", 3001,
		"The port of the HTTP server")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func (o *option) runE(cmd *cobra.Command, args []string) (err error) {
	var config *rest.Config
	if config, err = rest.InClusterConfig(); err != nil {
		return
	}

	var client *kubernetes.Clientset
	if client, err = kubernetes.NewForConfig(config); err != nil {
		return
	}

	http.HandleFunc("/api/v1/template.execute", plugin(&DefaultPluginExecutor{option: o}, client, ""))
	err = http.ListenAndServe(fmt.Sprintf(":%d", o.Port), nil)
	return
}

type option struct {
	Provider string
	Server   string
	Username string
	Token    string
	Port     int

	Owner       string
	Repo        string
	PR          string
	Status      string
	Target      string
	Label       string
	Description string
}

type DefaultPluginExecutor struct {
	option *option
}

type pluginOption struct {
	Option *option `json:"gogit-executor-plugin"`
}

func (e *DefaultPluginExecutor) Execute(args executor.ExecuteTemplateArgs) (resp executor.ExecuteTemplateResponse, err error) {
	p := args.Template.Plugin.Value
	fmt.Println("raw data", string(p))

	opt := &pluginOption{Option: e.option}
	if err = json.Unmarshal(p, opt); err != nil {
		return
	}

	fmt.Println("option is", *opt.Option)
	// TODO get more information from context
	repo := pkg.RepoInformation{
		Provider:    opt.Option.Provider,
		Server:      opt.Option.Server,
		Owner:       opt.Option.Owner,
		Repo:        opt.Option.Repo,
		Target:      opt.Option.Target,
		Username:    opt.Option.Username,
		Token:       opt.Option.Token,
		Status:      opt.Option.Status,
		Label:       opt.Option.Label,
		Description: opt.Option.Description,
	}
	if repo.PrNumber, err = strconv.Atoi(opt.Option.PR); err != nil {
		err = fmt.Errorf("wrong pull-request number, %v", err)
		return
	}
	if err = pkg.Reconcile(context.Background(), repo); err == nil {
		resp = executor.ExecuteTemplateResponse{
			Body: executor.ExecuteTemplateReply{
				Node: &wfv1.NodeResult{
					Phase:   wfv1.NodeSucceeded,
					Message: "success",
				},
			},
		}
		fmt.Println("send success")
	} else {
		fmt.Println("failed to send", err)
	}
	return
}

type PluginExecutor interface {
	// Execute commands based on the args provided from the workflow
	Execute(args executor.ExecuteTemplateArgs) (executor.ExecuteTemplateResponse, error)
}

var (
	ErrWrongContentType = errors.New("Content-Type header is not set to 'appliaction/json'")
	ErrReadingBody      = errors.New("Couldn't read request body")
	ErrMarshallingBody  = errors.New("Couldn't unmrashal request body")
	ErrExecutingPlugin  = errors.New("Error occured while executing plugin")
)

func plugin(p PluginExecutor, kubeClient kubernetes.Interface, namespace string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if header := req.Header.Get("Content-Type"); header != "application/json" {
			http.Error(w, ErrWrongContentType.Error(), http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, ErrReadingBody.Error(), http.StatusBadRequest)
			return
		}

		args := executor.ExecuteTemplateArgs{}
		if err := json.Unmarshal(body, &args); err != nil || args.Workflow == nil || args.Template == nil {
			http.Error(w, ErrMarshallingBody.Error(), http.StatusBadRequest)
			return
		}

		resp, err := p.Execute(args)
		if err != nil {
			fmt.Println("failed to execute plugin", err)
			http.Error(w, ErrExecutingPlugin.Error(), http.StatusInternalServerError)
			return
		}

		jsonResp, err := json.Marshal(resp.Body)
		if err != nil {
			fmt.Println("something went wrong", err)
			http.Error(w, "something went wrong", http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResp)
		return
	}
}
