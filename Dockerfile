ARG BUILD_BASE=golang:1.19.4
ARG FINAL_BASE=alpine:3.17
ARG VERSION=v0.0.1-default
ARG USER=artomator

# BUILD
FROM $BUILD_BASE as builder
WORKDIR /src/
WORKDIR /src/
COPY . /src/
ARG VERSION
ENV VERSION=$VERSION GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
    -ldflags="-w -s -X main.version=${VERSION} -extldflags '-static'" \
    -a -mod vendor -o server ./cmd/server/main.go

# RUN
FROM $FINAL_BASE
ARG VERSION
LABEL artomator.version="${VERSION}"
COPY --from=builder /src/server /app/
COPY --from=builder /src/bin /app/
WORKDIR /app
RUN apk add --update bash curl jq cosign ca-certificates python3
# gcloud
ENV CLOUDSDK_INSTALL_DIR /gcloud/
RUN curl -sSL https://sdk.cloud.google.com | bash
ENV PATH=/gcloud/:$PATH
# anchore tools 
RUN curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh \
    | sh -s -- -b /usr/local/bin
# aquasecurity
RUN curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh \
    | sh -s -- -b /usr/local/bin
# automator
ENTRYPOINT ["./server"]