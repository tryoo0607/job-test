############## BUILD ##############

# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder
WORKDIR /app

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64


# deps 캐시
COPY go.mod go.sum ./
RUN go mod download

# 앱 소스
COPY . .
COPY assets/inputs/ /data/inputs/

RUN go build -trimpath -ldflags="-s -w" -o /job-test-build .

############## RUN ##############
# FROM scratch
FROM busybox:1.36

WORKDIR /
COPY --from=builder /job-test-build /job-test
COPY --from=builder /data/inputs /data/inputs

ENTRYPOINT ["/job-test"]
