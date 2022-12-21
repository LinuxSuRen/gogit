package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/linuxsuren/gogit/pkg"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"

	"github.com/argoproj/argo-workflows/v3/pkg/plugins/executor"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	PR          int
	Status      string
	Target      string
	Label       string
	Description string
}

type DefaultPluginExecutor struct {
	option *option
}

func (e *DefaultPluginExecutor) Execute(args executor.ExecuteTemplateArgs) (resp executor.ExecuteTemplateResponse, err error) {
	p := args.Template.Plugin.Value

	opt := e.option
	if err = json.Unmarshal(p, opt); err != nil {
		return
	}

	// TODO get more information from context
	err = pkg.Reconcile(context.Background(), pkg.RepoInformation{
		Provider:    opt.Provider,
		Server:      opt.Server,
		Owner:       opt.Owner,
		Repo:        opt.Repo,
		PrNumber:    opt.PR,
		Target:      opt.Target,
		Username:    opt.Username,
		Token:       opt.Token,
		Status:      opt.Status,
		Label:       opt.Label,
		Description: opt.Description,
	})
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
			http.Error(w, ErrExecutingPlugin.Error(), http.StatusInternalServerError)
			return
		}

		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "something went wrong", http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResp)
		return
	}
}
