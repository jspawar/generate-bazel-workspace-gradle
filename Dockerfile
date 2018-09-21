FROM golang:1.10-alpine

COPY . /go/src/github.com/jspawar/generate-bazel-workspace-gradle/
WORKDIR /go/src/github.com/jspawar/generate-bazel-workspace-gradle/
RUN go test -v ./... -ginkgo.v -ginkgo.randomizeAllSpecs
