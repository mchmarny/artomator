# artomator (Artifact Registry Automator, naming is hard)

[Artifact Registry (AR)](https://cloud.google.com/artifact-registry) `artomator` automates the image signing, creation of [Software Bill of Materials (SBOM)](https://www.cisa.gov/sbom), and vulnerability scanning. Using [image labels](https://docs.docker.com/config/labels-custom-metadata/), you can indicate to `artomator` the type of processing you want it to perform on that image. For example:

```shell
docker build -t $IMAGE_TAG --label artomator-sbom=true --label artomator-vuln=true .
```

When that image is pushed to AR, `artomator` will automatically generate both signed SBOM and vulnerability report and add these as attestations to the image.

![](images/reg.png)

> The `artomator` service in Cloud Run scales to 0 so there is no additional cost when no new images are bing published. 

## how it works

Artifact Registry will automatically published [registry events](https://cloud.google.com/artifact-registry/docs/configure-notifications) if there is a [PubSub](https://cloud.google.com/pubsub/docs/overview) topic named `gcr` in the same project. `artomator` creates [Cloud Run](https://cloud.google.com/run) services which subscribes to that topic and processes any image that has at least one of the `artomator-sbom=true` or `artomator-vuln=true` labels based on the image digest.

> To prevent reprocessing the same images multiple times, `artomator` uses redis store to cache the processed image hashes.

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

## verify 

To verify the attestation for `artomator` processed images you will need the KMS key name that was used to sign that image. You retrieve it using the following command:

```shell
export SIGN_KEY=$(gcloud kms keys describe artomator-signer \
  --project $PROJECT_ID \
  --location $REGION \
  --keyring artomator \
  --format json | jq --raw-output '.name')
```

You can check the key like this: 

```shell
echo $SIGN_KEY
```

It should look something like this

```shell
projects/$PROJECT_ID/locations/$REGION/keyRings/artomator/cryptoKeys/artomator-signer/cryptoKeyVersions/1
```

Once have the signing key, you can verify any image that was processed by `artomator` like this:

> Note, the `$IMAGE_SHA` has to be the fully qualified image URI with the SHA. For example `us-west1-docker.pkg.dev/cloudy-demos/artomator/tester@sha256:59d5b8eb5525307dde52aa51382676e74240bb79eb92a67a1f2a760382a21d78`

```shell
cosign verify-attestation --type=spdx  --key "gcpkms://${SIGN_KEY}" $IMAGE_SHA \
    | jq -r .payload | base64 -d | jq -r .predicateType
```

> Note, you can check the attestation for either of the two types that `artomator` creates by changing the `--type` flag in the above command to either `spdx` (SBOM), `vuln` which is the vulnerability report

The result should look something like this: 

```shell
Verification for us-west1-docker.pkg.dev/cloudy-demos/artomator/tester@sha256:59d5b8eb5525307dde52aa51382676e74240bb79eb92a67a1f2a760382a21d78 --
The following checks were performed on each of these signatures:
  - The cosign claims were validated
  - The signatures were verified against the specified public key
https://spdx.dev/Document
```

To save any of these artifacts locally: 

```shell
cosign verify-attestation --type=spdx  --key "gcpkms://${SIGN_KEY}" $IMAGE_SHA \
    | jq -r .payload | base64 -d > sbom.spdx.json
```

## cleanup

To delete all created resources run: 

```shell
bin/cleanup
```

## todo

1. Save SBOM and vulnerability reports to GCS bucket 
1. Add UI to query images metadata (e.g. list packages, base images, or vulnerabilities over time)

## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.