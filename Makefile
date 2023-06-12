build:
	CGO_ENABLED=0 go build -ldflags "-w -s" -o bin/gogit
plugin-build:
	cd cmd/argoworkflow && CGO_ENABLED=0 go build -ldflags "-w -s" -o bin/gogit-executor-plugin
copy: build
	cp bin/gogit /usr/local/bin
test:
	go test ./... -coverprofile coverage.out
pre-commit: test build plugin-build
goreleaser:
	goreleaser build --snapshot --rm-dist
image:
	docker build . -t ghcr.io/linuxsuren/gogit
build-workflow-executor-gogit:
	cd cmd/argoworkflow && CGO_ENABLED=0 go build -ldflags "-w -s" -o ../../bin/workflow-executor-gogit
image-workflow-executor-gogit:
	cd cmd/argoworkflow && docker build . -t ghcr.io/linuxsuren/workflow-executor-gogit:dev --build-arg GOPROXY=https://goproxy.io,direct
push-image-workflow-executor-gogit: image-workflow-executor-gogit
	docker push ghcr.io/linuxsuren/workflow-executor-gogit:dev
test-workflow-executor-gogit:
	cd cmd/argoworkflow && go test ./...
