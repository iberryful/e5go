FROM golang:alpine AS builder
WORKDIR /app/
COPY . . 
RUN apk --update --no-cache --no-progress add git \
    && go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go build -o e5gobot main.go \
    && rm .gitignore build.sh Dockerfile e5go.service go.mod go.sum main.go README.md

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/ /app/
EXPOSE 3000

ENTRYPOINT ["/app/e5gobot"]