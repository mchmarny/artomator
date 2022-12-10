# artomator (Artifact Registry Automator, naming is hard)

[Artifact Registry (AR)](https://cloud.google.com/artifact-registry) `artomator`. When deployed, it creates Cloud Run services which subscribes to AR events. Whenever new image is published to any registry in specified project, this service will:

* Sign that image based on its SHA using a specific KMS key
* Generate [Software Bill of Materials (SBOM)](https://www.cisa.gov/sbom) for that image in ([JSON SPDX format](https://github.com/spdx/spdx-spec/blob/v2.2/schemas/spdx-schema.json))
* Scan that image for vulnerabilities using the extracted SBOM
* and, create a verifiable attestations for both the SBOM and vulnerability report on that image

![](images/reg.png)

## setup

Enable required APIs, create registry, and configure KMS key:

```shell
bin/setup
```

Create `artomator` image, sign it, generate its own SBOM, create its vulnerability report, publish that image to registry, and run attestation validation to make sure the image is ready for use:

```shell
bin/image
```

Finally, create the PubSub topic, subscription to that topic, and Cloud Run service to process the registry events: 

> Note, the created Cloud Run service requires `roles/run.invoker` roles so only the PubSub push subscription will be allowed to invoke that service. 

```shell
bin/deploy
```

## test 

To test `artomator`, use the provided test with ["hello" Dockerfile](test/Dockerfile): 

```shell
test/run
```

## cleanup

To delete all created resources run: 

```shell
bin/cleanup
```

## todo

1. Persist sha to prevent processing the same one multiple times 
1. Check for public key presence before invoking KMS API to workaround quota limitations
1. Save SBOM and vulnerability reports to GCS bucket 
1. Add UI to query images metadata (e.g. list packages, vulns over time, base images)

## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.