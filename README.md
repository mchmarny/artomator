# artomator

[![Go Report Card](https://goreportcard.com/badge/github.com/mchmarny/artomator)](https://goreportcard.com/report/github.com/mchmarny/artomator) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mchmarny/artomator) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/gojp/goreportcard/blob/master/LICENSE)


`artomator` (aka Artifact Registry Automator) automates the creation of [Software Bill of Materials (SBOM)](https://www.cisa.gov/sbom) with Binary Authorization attestation for container images in [Artifact Registry (AR)](https://cloud.google.com/artifact-registry). `artomator` will automatically add SBOM attestations to any image pushed to registry with the `artomator-sbom` [label](https://docs.docker.com/config/labels-custom-metadata/).

```shell
docker build -t $TAG --label artomator-sbom=spdx .
```

 The value of the label dictates SBOM format. The two supported formats are `cyclonedx` and `spdx`). `artomator` also creates [Binary Authorization](https://cloud.google.com/binary-authorization) attestation to support project or cluster levels policies.

![](images/flow.png)

## how it works

1. Whenever an image is published to the Artifact Registry 
2. A [registry event](https://cloud.google.com/artifact-registry/docs/configure-notifications) is automatically published onto [PubSub](https://cloud.google.com/pubsub/docs/overview) topic named `gcr`
3. PubSub subscription pushes that event to `artomator` service in [Cloud Run](https://cloud.google.com/run) with operation type and the image digest
4. If the operation type is `INSERT`, the `artomator` service retrieves metadata for that image from registry and check its labels
5. If the image includes `artomator-sbom` label, the service signs that image using KMS key
6. And creates new attestation based on the type of the label to the image in the registry (e.g. `spdx`)
7. If [GCS bucket](https://cloud.google.com/storage) is configured, `artomator` will also save the generated artifacts to that bucket
8. On successful completion, `artomator` also creates Binary Authorization attestation using `artomator-attestor` with associated KMS key
9. Finally `artomator` also stores the processed image digests in a [Redis store](https://cloud.google.com/memorystore) to avoid re-processing the same image again

> Technically, adding attestation to an image creates yet another event, and could cause recursion. To prevent this and to allow `artomator` to scale to multiple instances the Redis-based cache is used which caches the processed digests for 72 hrs.

To processes images, `artomator` uses a few open source projects:

* [cosign](https://github.com/sigstore/cosign) for image signing and verification
* [syft](https://github.com/anchore/syft) for SBOM generation 
* [trivy](https://github.com/aquasecurity/trivy) for vulnerability scans 
* [jq](https://stedolan.github.io/jq/) for JSON operations 

## artifacts 

In addition to attaching attestations to image in Artifact Registry and the Binary Authorization note, `artomator` also saves all the generated reports in GCS bucket (for example [sbom.json](tests/sbom.json)). To make these names predictable, `artomator` prefixes them with the image SHA. For example, if the image digest is:

```shell
us-west1-docker.pkg.dev/s3cme1/artomator/tester@sha256:acaccb6c8f975ee7df7f46468fae28fb5014cf02c2835d2dc37bf6961e648838
```

then the list of artifacts in the registry for that image will be: 

* acaccb6c8f975ee7df7f46468fae28fb5014cf02c2835d2dc37bf6961e648838-sbom.json
* acaccb6c8f975ee7df7f46468fae28fb5014cf02c2835d2dc37bf6961e648838-meta.json

where:

* `-sbom.json` is SPDX 2.3 formatted SBOM file
* `-meta.json` is the image metadata in the registry as it was when the image was processed

## deployment 

The prerequisites to deploy `artomator` include: 

* [Terraform CLI](https://www.terraform.io/downloads)
* [GCP Project](https://cloud.google.com/resource-manager/docs/creating-managing-projects)
* [gcloud CLI](https://cloud.google.com/sdk/gcloud)
  
To deploy the prebuilt `artomator`, first clone this repo:

```shell
git clone git@github.com:mchmarny/artomator.git
```

Then navigate to the `deployment` directory inside of that cloned repo:

```shell
cd artomator/deployment
```

Next, authenticate to GCP:

```shell
gcloud auth application-default login
```

Initialize Terraform: 

```shell
terraform init
```

> Note, this flow uses the default, local terraform state. Make sure you do not check the state files into your source control (see `.gitignore`), or consider using persistent state provider like GCS.


When done, apply the Terraform configuration:

```shell
terraform apply
```

When promoted, provide requested variables:

* `project_id` is the GCP project ID (not the name)
* `location` is GCP region to deploy to

When completed, this will output the configured resource information. 

## test deployment

To test the deployed `artomator`, use any valid Dockerfile you already have:

```shell
docker build -t $TEST_IMAGE_TAG --label artomator-sbom=spdx .
docker push $TEST_IMAGE_TAG
```

## cleanup

To clean all the resources provisioned by this setup run: 

```shell
terraform destroy
```

> Note, this does not remove the created KMS resources.

## disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.
