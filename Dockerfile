# Build Go Server Binary
FROM golang:1.14
LABEL MAINTAINER KeisukeYamashita<19yamashita15@gmail.com>

ENV GO111MODULE on

ARG SERVICE_NAME
ARG VERSION

WORKDIR /project

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build main.go

FROM alpine:latest
COPY --from=0 /project/main /bin/slackduty
ENTRYPOINT ["/bin/slackduty"]
