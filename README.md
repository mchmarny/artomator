# artomator (Artifact Registry Automator, naming is hard)

[Artifact Registry (AR)](https://cloud.google.com/artifact-registry) `artomator` automates the image signing, creation of [Software Bill of Materials (SBOM)](https://www.cisa.gov/sbom), and vulnerability scanning. Using image labels, you can indicate to `artomator` the type of processing you want it to perform on that image. For example:

```shell
docker build -t $IMAGE_TAG --label sbom=true --label vuln=true .
```

When that image is pushed to AR, `artomator` will automatically generate both signed SBOM and vulnerability report and add these as attestations to the image.

![](images/reg.png)

> The `artomator` service in Cloud Run scales to 0 so there is no additional cost when no new images are bing published. 

## how it works

Artifact Registry will automatically published [registry events](https://cloud.google.com/artifact-registry/docs/configure-notifications) if there is a [PubSub](https://cloud.google.com/pubsub/docs/overview) topic named `gcr` in the same project. `artomator` creates [Cloud Run](https://cloud.google.com/run) services which subscribes to that topic and processes any image that has at least one of the `sbom=true` or `vuln=true` labels based on the image digest.

`artomator` uses following OSS technologies: 

* [cosign](https://github.com/sigstore/cosign) with [GCP KMS](https://cloud.google.com/security-key-management) for image signing and verification
* [syft](https://github.com/anchore/syft) for SBOM generation 
* [grype](https://github.com/anchore/grype) for vulnerability scans 
* [jq](https://stedolan.github.io/jq/) for JSON operations 

## deployment 

To deploy the prebuilt `artomator` image with all the dependencies run:

```shell
bin/deploy-all
```

This will:

* Enable required APIs
* Create artifact registry (`artomator`)
* Configure KMS key (`keyRings/artomator/cryptoKeys/artomator-signer`)
* PubSub topic (`gcr`) and subscription to that topic (`gcr-sub`)
* Deploy Cloud Run service (`artomator`) to process the registry events

## build your own

To build the `artomator` image yourself in your own project, first, enable required APIs, create registry, and configure KMS key:

```shell
bin/setup
```

Then, build the `artomator` image locally, sign it, generate its own SBOM with vulnerability report, publish that image to your own registry (created in setup), and run attestation validation to make sure the image is ready for use:

```shell
bin/image
```

Finally, create the PubSub topic with push subscription, and deploy Cloud Run service to process the registry events: 

> Note, the created Cloud Run service requires `roles/run.invoker` roles so only the PubSub push subscription will be allowed to invoke that service. 

```shell
bin/deploy
```

## test 

To test `artomator`, use the provided test with ["hello" Dockerfile](tests/Dockerfile): 

```shell
tests/run
```

## cleanup

To delete all created resources run: 

```shell
bin/cleanup
```

## todo

1. Persist sha to prevent processing the same one multiple times (in project context)
1. Check for public key presence before invoking KMS API to workaround quota limitations
1. Save SBOM and vulnerability reports to GCS bucket 
1. Add UI to query images metadata (e.g. list packages, vulns over time, base images)

## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.