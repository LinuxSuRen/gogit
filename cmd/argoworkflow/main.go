package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	wfclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"github.com/argoproj/argo-workflows/v3/pkg/plugins/executor"
	"github.com/linuxsuren/gogit/argoworkflow/template"
	"github.com/linuxsuren/gogit/pkg"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	flags.StringVarP(&opt.Target, "target", "", "http://argo.argo-server.svc:2746",
		"The root URL of Argo Workflows UI")
	flags.IntVarP(&opt.Port, "port", "", 3001,
		"The port of the HTTP server")
	flags.BoolVarP(&opt.CreateComment, "create-comment", "", false, "Indicate if want to create a status comment")
	flags.StringVarP(&opt.CommentTemplate, "comment-template", "", "", "The template of the comment")
	flags.StringVarP(&opt.CommentIdentity, "comment-identity", "", pkg.CommentEndMarker, "The identity for matching exiting comment")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func (o *option) runE(cmd *cobra.Command, args []string) (err error) {
	var config *rest.Config
	if config, err = rest.InClusterConfig(); err != nil {
		return
	}
	client := wfclientset.NewForConfigOrDie(config)

	http.HandleFunc("/api/v1/template.execute", plugin(&DefaultPluginExecutor{option: o}, client))
	err = http.ListenAndServe(fmt.Sprintf(":%d", o.Port), nil)
	return
}

type option struct {
	Provider string
	Server   string
	Username string
	Token    string
	Port     int

	CreateComment   bool
	CommentTemplate string
	CommentIdentity string

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

func (e *DefaultPluginExecutor) Execute(args executor.ExecuteTemplateArgs, wf *wfv1.Workflow) (
	resp executor.ExecuteTemplateResponse, err error) {
	ctx := context.Background()
	var name string
	if wf.Spec.WorkflowTemplateRef != nil {
		name = wf.Spec.WorkflowTemplateRef.Name
	}
	wf.Status.Phase = wfv1.WorkflowPhase(wf.Status.Nodes[wf.Name].Phase)
	status := wf.Status

	p := args.Template.Plugin.Value

	opt := &pluginOption{Option: e.option}
	if err = json.Unmarshal(p, opt); err != nil {
		return
	}
	fmt.Println("option is", *opt.Option)

	targetAddress := fmt.Sprintf("%s/workflows/%s/%s",
		opt.Option.Target,
		args.Workflow.ObjectMeta.Namespace,
		args.Workflow.ObjectMeta.Name)
	repo := pkg.RepoInformation{
		Provider:    opt.Option.Provider,
		Server:      opt.Option.Server,
		Owner:       opt.Option.Owner,
		Repo:        opt.Option.Repo,
		Target:      targetAddress,
		Username:    opt.Option.Username,
		Token:       opt.Option.Token,
		Status:      opt.Option.Status,
		Label:       EmptyThen(opt.Option.Label, name),
		Description: EmptyThen(opt.Option.Description, status.Message),
	}
	if repo.Status == "" {
		switch status.Phase {
		case wfv1.WorkflowSucceeded:
			// from Argo Workflow
			repo.Status = "success"
		case wfv1.WorkflowFailed:
			repo.Status = "failure"
		default:
			repo.Status = strings.ToLower(string(status.Phase))
		}
	}
	if repo.PrNumber, err = strconv.Atoi(opt.Option.PR); err != nil {
		err = fmt.Errorf("wrong pull-request number, %v", err)
		return
	}

	fmt.Println("send status", repo)
	var nodeResult *wfv1.NodeResult
	if err = pkg.CreateStatus(ctx, repo); err == nil {
		nodeResult = &wfv1.NodeResult{
			Phase:   wfv1.NodeSucceeded,
			Message: "success",
		}
		fmt.Println("send status success")
	} else {
		fmt.Println("failed to send status", err)
	}

	if err == nil && opt.Option.CreateComment {
		fmt.Println("start to create comment")
		tplText := EmptyThen(opt.Option.CommentTemplate, template.CommentTemplate)

		// find useless nodes
		var toRemoves []string
		for key, val := range wf.Status.Nodes {
			if strings.HasSuffix(val.Name, ".onExit") || strings.Contains(val.Name, ".hooks.") {
				toRemoves = append(toRemoves, key)
			}
		}
		// remove useless nodes
		delete(wf.Status.Nodes, wf.Name)
		for _, key := range toRemoves {
			delete(wf.Status.Nodes, key)
		}

		// put the workflow link into annotations
		if wf.Annotations == nil {
			wf.Annotations = map[string]string{}
		}

		var templatePath string
		if wf.Spec.WorkflowTemplateRef.ClusterScope {
			templatePath = "cluster-workflow-templates"
		} else {
			templatePath = "workflow-templates"
		}
		targetTemplateAddress := fmt.Sprintf("%s/%s/%s/%s",
			opt.Option.Target,
			templatePath,
			args.Workflow.ObjectMeta.Namespace,
			wf.Spec.WorkflowTemplateRef.Name)
		wf.Annotations["workflow.link"] = targetAddress
		wf.Annotations["workflow.templatelink"] = targetTemplateAddress

		var message string
		message, err = template.RenderTemplate(tplText, wf)
		if err == nil {
			err = pkg.CreateComment(ctx, repo, message, opt.Option.CommentIdentity)
		} else {
			err = fmt.Errorf("failed to render comment template: %v", err)
		}

		if err != nil {
			fmt.Println("failed to create comment", err)
		}
	}

	if err != nil {
		nodeResult = &wfv1.NodeResult{
			Phase:   wfv1.NodeFailed,
			Message: err.Error(),
		}
	}

	resp = executor.ExecuteTemplateResponse{
		Body: executor.ExecuteTemplateReply{
			Node: nodeResult,
		},
	}
	return
}

type PluginExecutor interface {
	// Execute commands based on the args provided from the workflow
	Execute(args executor.ExecuteTemplateArgs, wf *wfv1.Workflow) (executor.ExecuteTemplateResponse, error)
}

var (
	ErrWrongContentType = errors.New("Content-Type header is not set to 'appliaction/json'")
	ErrReadingBody      = errors.New("Couldn't read request body")
	ErrMarshallingBody  = errors.New("Couldn't unmrashal request body")
)

func plugin(p PluginExecutor, client *wfclientset.Clientset) func(w http.ResponseWriter, req *http.Request) {
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

		fmt.Println(string(body))
		args := executor.ExecuteTemplateArgs{}
		if err := json.Unmarshal(body, &args); err != nil || args.Workflow == nil || args.Template == nil {
			http.Error(w, ErrMarshallingBody.Error(), http.StatusBadRequest)
			return
		}

		go func(c *wfclientset.Clientset, args executor.ExecuteTemplateArgs) {
			time.Sleep(3 * time.Second)
			wfName := args.Workflow.ObjectMeta.Name
			wfNamespace := args.Workflow.ObjectMeta.Namespace

			// find the Workflow
			var workflow *wfv1.Workflow
			if workflow, err = client.ArgoprojV1alpha1().Workflows(wfNamespace).Get(
				context.Background(),
				wfName,
				v1.GetOptions{}); err != nil {
				fmt.Println("failed to find workflow", wfName, wfNamespace, err)
				return
			}

			_, _ = p.Execute(args, workflow)
		}(client, args)

		jsonResp, err := json.Marshal(executor.ExecuteTemplateReply{
			Node: &wfv1.NodeResult{
				Phase:   wfv1.NodeSucceeded,
				Message: "success",
			},
		})
		if err != nil {
			fmt.Println("something went wrong", err)
			http.Error(w, "something went wrong", http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonResp)
		return
	}
}

// EmptyThen return second if the first is empty
func EmptyThen(first, second string) string {
	if first == "" {
		return second
	} else {
		return first
	}
}
