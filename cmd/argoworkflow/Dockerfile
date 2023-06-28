FROM golang:1.18 as builder
ARG GOPROXY=direct

WORKDIR /workspace
COPY . .

RUN go mod download
RUN GOPROXY=${GOPROXY} CGO_ENABLED=0 go build -ldflags "-w -s" -o workflow-executor-gogit

FROM ubuntu:kinetic

COPY --from=builder /workspace/workflow-executor-gogit /usr/bin/workflow-executor-gogit
RUN apt update -y && apt install ca-certificates -y
CMD ["workflow-executor-gogit"]
