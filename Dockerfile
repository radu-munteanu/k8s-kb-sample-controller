# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/gitlab.com/radu-munteanu/k8s-kb-sample-controller
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager gitlab.com/radu-munteanu/k8s-kb-sample-controller/cmd/manager

# Copy the controller-manager into a thin image
FROM ubuntu:latest
WORKDIR /
COPY --from=builder /go/src/gitlab.com/radu-munteanu/k8s-kb-sample-controller/manager .
ENTRYPOINT ["/manager"]
