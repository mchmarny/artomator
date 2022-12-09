# BUILD
FROM golang:1.19.4 as builder

WORKDIR /src/

# copy
COPY go.* /src/
RUN go mod download
COPY main.go /src/

# runtime args
ARG VERSION=v0.0.1-default

# args to env vars
ENV VERSION=${VERSION} GO111MODULE=on

# build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
    -ldflags="-w -s -X main.version=${VERSION} -extldflags '-static'" \
    -a -mod readonly -v -o app

# RUN
FROM alpine:3.17

COPY --from=builder /src/app /app/
WORKDIR /app

# core packages + py for gcloud
RUN apk add --no-cache bash curl docker jq cosign ca-certificates python3 

# gcloud
RUN mkdir -p /builder && \
    wget -qO- https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz | tar zxv -C /builder && \
    /builder/google-cloud-sdk/install.sh --usage-reporting=false \
        --bash-completion=false \
        --disable-installation-options

# add gcloud to path 
ENV PATH=/builder/google-cloud-sdk/bin/:$PATH

# anchore tools 
RUN curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh \
    | sh -s -- -b /usr/local/bin
RUN curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh \
    | sh -s -- -b /usr/local/bin

# copy automator
COPY bin/artomator /usr/local/bin
ENTRYPOINT ["./app"]