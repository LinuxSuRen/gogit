build:
	CGO_ENABLE=0 go build -ldflags "-w -s" -o bin/gogit
copy: build
	cp bin/gogit /usr/local/bin
test:
	go test ./... -coverprofile coverage.out
goreleaser:
	goreleaser build --snapshot --rm-dist
image:
	docker build . -t ghcr.io/linuxsuren/gogit
build-workflow-executor-gogit:
	cd cmd/argoworkflow && CGO_ENABLE=0 go build -ldflags "-w -s" -o ../../bin/workflow-executor-gogit
image-workflow-executor-gogit:
	cd cmd/argoworkflow && docker build . -t ghcr.io/linuxsuren/workflow-executor-gogit
test-workflow-executor-gogit:
	cd cmd/argoworkflow && go test ./...
