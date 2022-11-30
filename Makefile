build:
	CGO_ENABLE=0 go build -ldflags "-w -s" -o bin/gogit
goreleaser:
	goreleaser build --snapshot --rm-dist
image:
	docker build . -t ghcr.io/linuxsuren/gogit
