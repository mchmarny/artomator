# Description: Cloud Run service

# List of roles that will be assigned to the runner service account
locals {
  # List of roles that will be assigned to the runner service account
  runner_roles = toset([
    "roles/artifactregistry.writer",
    "roles/binaryauthorization.attestorsViewer",
    "roles/cloudkms.cryptoKeyDecrypter",
    "roles/cloudkms.signerVerifier",
    "roles/cloudkms.viewer",
    "roles/containeranalysis.notes.attacher",
    "roles/containeranalysis.occurrences.editor",
    "roles/iam.serviceAccountTokenCreator",
    "roles/monitoring.metricWriter",
    "roles/run.invoker",
    "roles/storage.objectCreator",
    "roles/storage.objectViewer",
  ])
}

# Service Account under which the Cloud Run services will run
resource "google_service_account" "runner_service_account" {
  account_id   = "${var.name}-run-sa"
  display_name = "Cloud Run service account for ${var.name}"
}

# Role binding
resource "google_project_iam_member" "runner_role_bindings" {
  for_each = local.runner_roles
  project  = var.project_id
  role     = each.value
  member   = "serviceAccount:${google_service_account.runner_service_account.email}"
}

# Cloud Run service 
resource "google_cloud_run_service" "app" {
  name                       = var.name
  location                   = var.location
  project                    = var.project_id
  autogenerate_revision_name = true

  template {
    spec {
      containers {
        image = "${var.image}:${data.template_file.version.rendered}"

        ports {
          name           = "http1"
          container_port = 8080
        }
        resources {
          limits = {
            cpu    = "1000m"
            memory = "2Gi"
          }
        }
        env {
          name  = "ADDRESS"
          value = ":8080"
        }
        env {
          name  = "PROJECT_ID"
          value = var.project_id
        }
        env {
          name  = "SIGN_KEY"
          value = data.google_kms_crypto_key_version.version.name
        }
        env {
          name  = "REDIS_IP"
          value = google_redis_instance.cache.host
        }
        env {
          name  = "REDIS_PORT"
          value = google_redis_instance.cache.port
        }
        env {
          name  = "GCS_BUCKET"
          value = google_storage_bucket.artifact_store.name
        }
        env {
          name  = "ATTESTOR_ID"
          value = google_binary_authorization_attestor.attestor.id
        }
      }

      container_concurrency = 80
      timeout_seconds       = 900
      service_account_name  = google_service_account.runner_service_account.email
    }
    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale"         = "3"
        "run.googleapis.com/vpc-access-connector"  = google_vpc_access_connector.connector.name
        "run.googleapis.com/vpc-access-egress"     = "private-ranges-only"
        "run.googleapis.com/execution-environment" = "gen2"
      }
      labels = {
        "run.googleapis.com/startupProbeType"      = "Default"
      }
    }
  }

  metadata {
    annotations = {
      "run.googleapis.com/ingress"     = "all"
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  lifecycle {
    ignore_changes = [
      metadata.0.annotations["run.googleapis.com/operation-id"],
    ]
  }
}

# IAM member to grant access to the Cloud Run service
resource "google_cloud_run_service_iam_member" "app-access" {
  location = google_cloud_run_service.app.location
  project  = google_cloud_run_service.app.project
  service  = google_cloud_run_service.app.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.runner_service_account.email}"
}
