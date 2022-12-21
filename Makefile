build:
	CGO_ENABLE=0 go build -ldflags "-w -s" -o bin/gogit
goreleaser:
	goreleaser build --snapshot --rm-dist
image:
	docker build . -t ghcr.io/linuxsuren/gogit
build-workflow-executor-gogit:
	cd cmd/argoworkflow && CGO_ENABLE=0 go build -ldflags "-w -s" -o ../../bin/workflow-executor-gogit
image-workflow-executor-gogit:
	cd cmd/argoworkflow && docker build . -t ghcr.io/linuxsuren/workflow-executor-gogit
