# artomator

[Artifact Registry](https://cloud.google.com/artifact-registry) `artomator`. When deployed, it creates Cloud Run services which subscribes to registry events. When the event indicates a new image has been pushed (`INSERT`) to the registry, the service will:

* Sign that image based on its SHA using KMS key
* Generate Software Bill of Materials (SBOM) for that image in JSON format ([SPDX schema ](https://github.com/spdx/spdx-spec/blob/v2.2/schemas/spdx-schema.json))
* Scan that image for vulnerabilities 
* and, create attestations for both the SBOM and vulnerability report on the image

## setup

Enable required APIs, create registry, and configure KMS key:

```shell
bin/setup
```

Create `artomator` image, sign it, generate its own SBOM, create its vulnerability report, publish that image to registry, and run attestation validation to make sure the image is ready for use:

```shell
bin/image
```

Finally, create the PubSub topic, subscription, and Cloud Run service to process the registry events: 

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

## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.