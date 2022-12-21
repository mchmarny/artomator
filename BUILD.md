# build your own

In most cases you can just follow the [deployment](README.md#deployment) instruction which use pre-build `artomator` image. This document will cover manual build and deployment into your own project. 

## pre-requisites 

* [gcloud](https://cloud.google.com/sdk/docs/install) with existing GCP project 
* [cosign](https://github.com/sigstore/cosign)
* [syft](https://github.com/anchore/syft)
* [trivy](https://github.com/aquasecurity/trivy)
* [jq](https://stedolan.github.io/jq/)

## setup 

First, enable required APIs, create registry, and configure KMS key:

```shell
tools/setup
```

## build 

Next, build the `artomator` image locally, sign it, generate its own SBOM with vulnerability report, publish that image to your own registry (created in setup), and run attestation validation to make sure the image is ready for use:

```shell
tools/build
```

Finally, create the PubSub topic with push subscription, and deploy Cloud Run service to process the registry events: 

> Note, the created Cloud Run service requires `roles/run.invoker` roles so only the PubSub push subscription will be allowed to invoke that service. 

```shell
tools/deploy
```

## disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.