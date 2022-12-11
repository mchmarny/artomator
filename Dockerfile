ARG BUILD_BASE=golang:1.19.4
ARG FINAL_BASE=alpine:3.17

# BUILD
FROM $BUILD_BASE as builder
WORKDIR /src/
WORKDIR /src/
COPY . /src/
ARG VERSION=v0.0.1-default
ENV VERSION=${VERSION} GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
    -ldflags="-w -s -X main.version=${VERSION} -extldflags '-static'" \
    -a -mod vendor -o app

# RUN
FROM $FINAL_BASE
COPY --from=builder /src/app /app/
WORKDIR /app
RUN apk add --update bash curl jq cosign ca-certificates python3
# gcloud
RUN mkdir -p /builder && \
    wget -qO- https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz | tar zxv -C /builder && \
    /builder/google-cloud-sdk/install.sh --usage-reporting=false \
        --bash-completion=false \
        --disable-installation-options
ENV PATH=/builder/google-cloud-sdk/bin/:$PATH
# anchore tools 
RUN curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh \
    | sh -s -- -b /usr/local/bin
RUN curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh \
    | sh -s -- -b /usr/local/bin
# automator
COPY artomator /usr/local/bin
ENTRYPOINT ["./app"]