FROM golang:1.18 as builder
ARG GOPROXY=direct

WORKDIR /workspace
COPY . .

RUN go mod download
RUN GOPROXY=${GOPROXY} CGO_ENABLED=0 go build -ldflags "-w -s" -o workflow-executor-gogit

FROM alpine:3.10

LABEL "repository"="https://github.com/linuxsuren/gogit"
LABEL "homepage"="https://github.com/linuxsuren/gogit"
LABEL "maintainer"="Rick"
LABEL "Name"="A tool for sending build status to git providers"

COPY --from=builder /workspace/workflow-executor-gogit /usr/bin/workflow-executor-gogit
RUN apk add ca-certificates

CMD ["workflow-executor-gogit"]
