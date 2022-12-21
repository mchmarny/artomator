resource "google_pubsub_topic" "gcr_topic" {
  name                       = "gcr"
  message_retention_duration = "86600s"
}

resource "google_pubsub_subscription" "gcr_sub" {
  name  = "${var.name}-gcr-sub"
  topic = google_pubsub_topic.gcr_topic.name
  ack_deadline_seconds = 600

  push_config {
    push_endpoint = "${google_cloud_run_service.app.status[0].url}/event"

    attributes = {
      x-goog-version = "v1"
    }

    oidc_token {
      service_account_email = google_service_account.runner_service_account.email
      audience              = "${google_cloud_run_service.app.status[0].url}/event"
    }
  }
}

resource "google_project_iam_member" "pubsub_token_creator" {
  project = data.google_project.project.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}
