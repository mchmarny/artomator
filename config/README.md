# oidc-for-gcp-using-terraform

Setup Github Workflow OpenID Connect in GCP for this repo using Terraform

## Prerequisites 

The prerequisites to executing this setup include: 

* [Terraform CLI](https://www.terraform.io/downloads)
* [GCP Project](https://cloud.google.com/resource-manager/docs/creating-managing-projects)
* [gcloud CLI](https://cloud.google.com/sdk/gcloud)
  
> Good how-to on using terraform with GCP is located [here](https://cloud.google.com/community/tutorials/getting-started-on-gcp-with-terraform).

## One-time Setup 

> Make sure to authenticate to GCP using `gcloud auth application-default login` if you haven't done it already.

```shell
terraform init
```

> Note, this flow uses the default, local terraform state. Make sure you do not check the state files into your source control (see `.gitignore`), or consider using persistent state provider like GCS.

## Executing Configuration 

Start by validating 

```shell
terraform validate
```

To configure Github Workflow OpenID Connect setup for GCP apply the cloned configuration:

```shell
terraform apply
```

When promoted, provide the 2 required variables:

* `project_id` is the GCP project ID (not the name) which you want to target from your GitHub Action. 
* `git_repo` is the username/repo combination in which you GitHub Actions will be executing

## What Included

You can review each one fo the `*.tf` files for content. When you confirm `yes` at the final prompt, the main artifacts created by this setup in the GCP project defined by the `project_id` variable include: 

* Enablement of the required GCP APIs
  * `servicecontrol.googleapis.com`
  * `containerregistry.googleapis.com`
  * `iam.googleapis.com`
  * `iamcredentials.googleapis.com`
  * `servicemanagement.googleapis.com`
  * `storage-api.googleapis.com`
* Creation of `github-actions-user` service account which the GitHub Action will impersonate when publishing images into GCR, and binding that account to the two required role:
  * `roles/storage.objectCreator`
  * `roles/storage.objectViewer`
* Creation of the workload identity pool: `github-pool`, and GitHub repo-level pool provider: `github-provider`
* Finally, creation of the IAM policy bindings to the service account resources created by GitHub identify for the specific GitHub repository defined by the `git_repo` variable

## Repo Configuration

The result each execution of the above defined configuration will include 3 GitHub repo configuration properties:

* `PROJECT_ID` which is the project ID in which you setup the workload identity federation
* `SERVICE_ACCOUNT` which is the IAM service account your GitHub Action workflows will use to push images into GCR (e.g. `github-action-publisher@<project_id>.iam.gserviceaccount.com`)
* `IDENTITY_PROVIDER` which si the workflow identity provider ID you must use lng with the above service account to connect to GCP (e.g. `projects/<project_number>/locations/global/workloadIdentityPools/github-pool/providers/github-provider`)

> Depending on your tolerance, you may be OK using all 3 of these parameters in your GitHub Actions workflow in plain-text. In most cases, however, you will probably create GitHub[secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets) in your repository to inject them into your workflow at runtime. 

## GitHub Workflow Configuration 

With the Workload Identity Federation configured yur workflow can now establish delegated trust relationship to the narrowly scoped set of permissions in GCP. The [google-github-actions/auth](https://github.com/google-github-actions/auth) includes many examples using `gcloud` in your workflow. 

In this post I'm going to focus on [Go](https://go.dev/)-specific configuration using [ko](https://github.com/google/ko), (a super simple and fast container image builder for Go apps) to build and publishing images into [GCR](https://cloud.google.com/container-registry). The full workflow is available [here](https://github.com/mchmarny/restme/blob/main/.github/workflows/image-on-tag.yaml). The key steps include: 

### Push Job

First, in order create OIDC tokens, the GitHub Actions will need additional permissions. In addition to regular `content` read, the workflow will also `id-token` write. 

```yaml
jobs:
  push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write
    steps:
```

### GCP Authentication

In order to push images to GCR, the workflow will need to first authenticate to GCP. Google has an action just for that that can be configured to generate OAuth 2.0 Access Token. To do this you will need to set the `token_format` to `access_token`. Additionally, this step will use the workload identity provider and service account secrets we configured above:

```yaml
    - id: auth
      name: Get GCP token
      uses: google-github-actions/auth@v0.5.0
      with:
          token_format: "access_token"
          workload_identity_provider: ${{ secrets.IDENTITY_PROVIDER }}
          service_account: ${{ secrets.SERVICE_ACCOUNT }}
```

### Install And Login Ko

Ko is the fastest way of creating container images in Go without Docker. All we need to do is install it and login to GCR with the access token created by the `auth` step above:

```yaml
    - name: Install Ko
      uses: imjasonh/setup-ko@v0.4
      with:
        version: tip
        
    - name: Login With ko
      run: |
        ko login gcr.io --username=oauth2accesstoken --password=${{ steps.auth.outputs.access_token }}
```

### Publish Image

With ko logged in, now you can build and publish the image. A few things to highlight here. `ko build` (pka `publish`) will build and publish container images from the given path. The `--image-refs` flag will output the digest of the published image to the provided file, and the `--bare` allows us to define the full image URL using the `KO_DOCKER_REPO` environment variable. 

In addition to this we will set the previously exported `RELEASE_VERSION` environment variable to both `version` field in the `main.go` file and set it as a tag on the image. 

```yaml
    - name: Publish Image
      run: |
        ko build ./cmd/ --image-refs ./image-digest --bare --tags ${{ env.RELEASE_VERSION }},latest
      env:
        KO_DOCKER_REPO: gcr.io/${{ secrets.PROJECT_ID }}/restme
        GOFLAGS: "-ldflags=-X=main.version=${{ env.RELEASE_VERSION }}"
```

### Sign Image

Once the image is published, we can also sign and verify the published image in GCR using [cosign](https://github.com/sigstore/cosign). 

```yaml
    - name: Install Cosign
      uses: sigstore/cosign-installer@main
      with:
        cosign-release: v1.4.1
```

The benefit of combining `ko` and `cosign` is that we can use the image digest output into a local file by `ko` by providing its path using `--force` flag in the `cosign sign` command. 

> With the v`1.4` release of cosign, you set th `COSIGN_EXPERIMENTAL` variable to push the data into GCR.

```yaml
    - name: Sign Image
      run: |
        cosign sign --force $(cat ./image-digest) 
      env:
        COSIGN_EXPERIMENTAL: 1
```


## Clean up

To clean all the resources provisioned by this setup run: 

```shell
terraform destroy
```