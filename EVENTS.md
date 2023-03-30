# Registry Events

Both Google Container Registry (GCR) and Artifact Registry (GAR) provide Pub/Sub notifications in form of Pub/Sub events:

* [Artifact Life-cycle](#container-life-cycle) (upload, tag, deletion) - GCR & GAR
* [Artifact Vulnerability](#artifact-vulnerabilities) (note, occurrence create or update) - GAR Only

## Container Life-cycle

> This event will fire regardless if you route images to GCR or AR as long as there is a topic name `gcr` in your project

Start by crating the `gcr` topic:

```shell
gcloud pubsub topics create gcr --project $PROJECT_ID
```

Next create subscription on that topic:

> The name doesn't matter

```shell
gcloud pubsub subscriptions create gcr-sub --project $PROJECT_ID --topic gcr
```

Now you can trigger the event: 

> Easiest way is to copy existing image using [crane](https://github.com/michaelsauter/crane). Make sure to substitute the below image tags. 

```shell
crane cp \
     $REGION-docker.pkg.dev/$PROJECT/repo1/image \
     $REGION-docker.pkg.dev/$PROJECT/repo2/image
```

Finally, list the events that were published about that artifact:

> Note, the content of events on the Pub/Sub topic is Base64 encoded so you will have to decode it using the `format` flag.

```shell
gcloud pubsub subscriptions pull gcr-sub --project $PROJECT_ID --auto-ack --limit 3 \
    --format="json(message.attributes, message.data.decode(\"base64\").decode(\"utf-8\"), message.messageId, message.publishTime)"
```

The content of each one of the events will look something like this: 

```json
{
    "message": {
        "data": {
            "action": "INSERT", 
            "digest": "us-west1-docker.pkg.dev/$PROJECT/repo/image@sha256:54bc0fead59f304f1727280c3b520aeea7b9e6fd405b7a6ee1dddc8d78044516", 
            "tag": "us-west1-docker.pkg.dev/$PROJECT/repo/image:latest"
        },
        "messageId": "7309198396944430",
        "publishTime": "2023-03-30T21:56:52.254Z"
    }
}
```

To clean up, you can delete the created subscription and topic:

```shell
gcloud pubsub subscriptions delete gcr-sub --project $PROJECT_ID
gcloud pubsub topics delete gcr --project $PROJECT_ID
```

## Artifact Vulnerabilities

If enabled, Artifact Analysis (aka Container Analysis) will create Pub/Sub events for each vulnerability found by automated AR scanning. For the most part, you can think of a `note` as data about vulnerability (e.g. `CVE`), and `occurrence` as the data that connects that `note` to a particular artifact (i.e. `digest`). 

> Note: these events do not fire for notes and occurrence created via Artifact Analysis API.

If you haven't already done so, start by enabling the Artifact Analysis and Container Scanning API 

```shell
gcloud services enable containeranalysis.googleapis.com --project $PROJECT_ID
gcloud services enable containerscanning.googleapis.com --project $PROJECT_ID
```

When enabled, Artifact Analysis API will automatically creates the Pub/Sub topics for both notes and occurrences. You can check if they exist using this command: 

```shell
gcloud pubsub topics list --project $PROJECT_ID
```

The results should looks something like this:

```shell
name: projects/$PROJECT/topics/container-analysis-occurrences-v1
name: projects/$PROJECT/topics/container-analysis-notes-v1
```

> If these topics do not exist, you can create them yourself using the `gcloud pubsub topics create` command

Next create subscription on the `occurrences` topic:

> The name doesn't matter. If you want, you can also create one for the notes topic.

```shell
gcloud pubsub subscriptions create vulns --project $PROJECT_ID --topic container-analysis-occurrences-v1
```

Now you will need to trigger the AA event:

> Again, simplest way to do that is to push an existing image using [crane](https://github.com/michaelsauter/crane).

```shell
crane cp \
     $REGION-docker.pkg.dev/$PROJECT/repo1/image \
     $REGION-docker.pkg.dev/$PROJECT/repo2/image
```

Finally list the vulnerabilities that were discovered:

```shell
gcloud pubsub subscriptions pull vulns --project $PROJECT_ID --auto-ack --limit 3 \
    --format="json(message.attributes, message.data.decode(\"base64\").decode(\"utf-8\"), message.messageId, message.publishTime)"
```

> If the above command returns an empty array (`[]`), give it a few seconds and rerun it. The length of the delay will depend on the size of your image and the number of vulnerabilities.

Each one of the vulnerabilities discovered by AA in your image will look something like this: 

```json
{
    "message": {
        "data": {
            "name": "projects/$PROJECT/occurrences/d2342144-8a7e-4f3c-b3ba-87ebbe3ac72d",
            "kind": "VULNERABILITY", 
            "notificationTime": "2023-03-30T23:09:28.471565Z"
        },
        "messageId": "7309675999864387",
        "publishTime": "2023-03-30T23:09:28.592Z"
    }
}
```

Once you have the id of the occurrence, you can use the Container Analysis REST API to get the details:

```shell
curl -Ss -H "Content-Type: application/json; charset=utf-8" \
     -H "Authorization: Bearer $(gcloud auth application-default print-access-token)" \
     https://containeranalysis.googleapis.com/v1/projects/$PROJECT/occurrences/d2342144-8a7e-4f3c-b3ba-87ebbe3ac72d
```

The response will look something like this:

```json
{
    "name": "projects/$PROJECT/occurrences/d2342144-8a7e-4f3c-b3ba-87ebbe3ac72d",
    "resourceUri": "https://us-west1-docker.pkg.dev/$PROJECT/$REPO/$IMAGE@sha256:5ffd30269c7bde2e29453bb9b8d3618814b7034e37aef299e3c071acbb565911",
    "noteName": "projects/$PROJECT/notes/CVE-2019-7577",
    "kind": "VULNERABILITY",
    "createTime": "2023-03-30T23:09:28.443028Z",
    "updateTime": "2023-03-30T23:09:28.443028Z",
    "vulnerability": {
        "severity": "MEDIUM",
        "cvssScore": 6.8,
        "packageIssue": [
            {
                "affectedCpeUri": "cpe:/o:canonical:ubuntu_linux:18.04",
                "affectedPackage": "libsdl2",
                "affectedVersion": {
                    "name": "2.0.8+dfsg1",
                    "revision": "1ubuntu1.18.04.5~18.04.1",
                    "kind": "NORMAL",
                    "fullName": "2.0.8+dfsg1-1ubuntu1.18.04.5~18.04.1"
                },
                "fixedCpeUri": "cpe:/o:canonical:ubuntu_linux:18.04",
                "fixedPackage": "libsdl2",
                "fixedVersion": {
                    "kind": "MAXIMUM"
                },
                "packageType": "OS",
                "effectiveSeverity": "LOW"
            }
        ],
        "shortDescription": "CVE-2019-7577",
        "longDescription": "NIST vectors: AV:N/AC:M/Au:N/C:P/I:P/A:P",
        "relatedUrls": [
            {
                "url": "http://people.ubuntu.com/~ubuntu-security/cve/CVE-2019-7577",
                "label": "More Info"
            }
        ],
        "effectiveSeverity": "LOW",
        "cvssv3": {
            "baseScore": 8.8,
            "exploitabilityScore": 2.8,
            "impactScore": 5.9,
            "attackVector": "ATTACK_VECTOR_NETWORK",
            "attackComplexity": "ATTACK_COMPLEXITY_LOW",
            "privilegesRequired": "PRIVILEGES_REQUIRED_NONE",
            "userInteraction": "USER_INTERACTION_REQUIRED",
            "scope": "SCOPE_UNCHANGED",
            "confidentialityImpact": "IMPACT_HIGH",
            "integrityImpact": "IMPACT_HIGH",
            "availabilityImpact": "IMPACT_HIGH"
        }
    }
}
```

To clean up, you can delete the created subscription and topic:

```shell
gcloud pubsub subscriptions delete vulns --project $PROJECT_ID
```







