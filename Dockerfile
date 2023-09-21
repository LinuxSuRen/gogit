FROM golang:1.18 as builder

WORKDIR /workspace
COPY cmd/ cmd
COPY pkg/ pkg
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go

RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags "-w -s" -o gogit

FROM alpine:3.10

LABEL "repository"="https://github.com/linuxsuren/gogit"
LABEL "homepage"="https://github.com/linuxsuren/gogit"
LABEL "maintainer"="Rick"
LABEL "Name"="A tool for sending build status to git providers"

COPY --from=builder /workspace/gogit /usr/bin/gogit
RUN apk add ca-certificates

ENTRYPOINT ["gogit"]
