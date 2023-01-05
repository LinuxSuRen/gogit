[![codecov](https://codecov.io/gh/LinuxSuRen/gogit/branch/master/graph/badge.svg?token=mnFyeD2IQ7)](https://codecov.io/gh/LinuxSuRen/gogit)

`gogit` could send the build status to different git providers. Such as:

* GitHub
* Gitlab (public or private)

## Usage

### Checkout to branch or PR
Ideally, `gogit` could checkout to your branch or PR in any kind of git repository.

You can run the following command in a git repository directory:

```shell
gogit checkout --pr 1
```

### Send status to Git Provider
Below is an example of sending build status to a private Gitlab server:

```shell
gogit status --provider gitlab \
  --server http://10.121.218.82:6080 \
  --repo yaml-readme \
  --pr 1 \
  --username linuxsuren \
  --token h-zez9CWzyzykbLoS53s
```

Or in the following use cases:

* [Tekton Task](https://hub.tekton.dev/tekton/task/gogit)

## Argo workflow Executor

Install as an Argo workflow executor plugin:

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gogit-executor-plugin
  namespace: default
---
apiVersion: v1
data:
  sidecar.automountServiceAccountToken: "true"
  sidecar.container: |
    args:
    - status
    - --provider
    - gitlab
    - --target
    - http://argo.argo-server.svc:2746                        # should be an external address
    image: ghcr.io/linuxsuren/workflow-executor-gogit:master
    command:
    - workflow-executor-gogit
    name: gogit-executor-plugin
    ports:
    - containerPort: 3001
    resources:
      limits:
        cpu: 500m
        memory: 128Mi
      requests:
        cpu: 250m
        memory: 64Mi
    securityContext:
      allowPrivilegeEscalation: true
      runAsNonRoot: true
      runAsUser: 65534
kind: ConfigMap
metadata:
  creationTimestamp: null
  labels:
    workflows.argoproj.io/configmap-type: ExecutorPlugin
  name: gogit-executor-plugin
  namespace: argo
```

then, create a WorkflowTemplate:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: WorkflowTemplate
metadata:
  name: plugin
  namespace: default
spec:
  entrypoint: main
  hooks:
    exit:
      template: status
    all:
      template: status
      expression: "true"
  templates:
  - container:
      args:
        - search
        - kubectl
      command:
        - hd
      image: ghcr.io/linuxsuren/hd:v0.0.70
    name: main
  - name: status
    plugin:
      gogit-executor-plugin:
        owner: linuxsuren
        repo: test
        pr: "3"
        label: test
```

## TODO
* Support more git providers

## Thanks
Thanks to these open source projects, they did a lot of important work.
* github.com/jenkins-x/go-scm
* github.com/spf13/cobra
