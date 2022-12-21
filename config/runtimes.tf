locals {
  # List of roles that will be assigned to the runner service account
  runner_roles = toset([
    "roles/artifactregistry.writer",
    "roles/browser",
    "roles/cloudkms.signerVerifier",
    "roles/cloudkms.viewer",
    "roles/iam.serviceAccountTokenCreator",
    "roles/monitoring.metricWriter",
    "roles/run.viewer",
    "roles/storage.objectCreator",
    "roles/storage.objectViewer",
    "roles/viewer",
  ])
}

# Service Account under which the Cloud Run services will run
resource "google_service_account" "runner_service_account" {
  account_id   = "${var.name}-run-sa"
  display_name = "Cloud Run service account for ${var.name}"
}

# Role binding to allow publisher to publish images
resource "google_project_iam_member" "runner_role_bindings" {
  for_each = local.runner_roles
  project  = var.project_id
  role     = each.value
  member   = "serviceAccount:${google_service_account.runner_service_account.email}"
}

# # Worker Cloud Run service 
# resource "google_cloud_run_service" "worker" {
#   name                       = var.name
#   location                   = var.region
#   project                    = var.project_id
#   autogenerate_revision_name = true

#   template {
#     spec {
#       containers {
#         image = "${var.location}-docker.pkg.dev/${var.project_id}/${var.name}/${var.name}:${data.template_file.version.rendered}"

#         ports {
#           name           = "http1"
#           container_port = 8080
#         }
#         resources {
#           limits = {
#             cpu    = "1000m"
#             memory = "2048Mi"
#           }
#         }
#         env {
#           name  = "ADDRESS"
#           value = ":8080"
#         }
#         env {
#           name  = "CONFIG"
#           value = "/secrets/${var.name}"
#         }
#       }
#       volumes {
#         name = "config-secret"
#         secret {
#           secret_name = google_secret_manager_secret.config_secret.secret_id
#           items {
#             key  = var.config_secret_version
#             path = var.name
#           }
#         }
#       }

#       container_concurrency = 80
#       timeout_seconds       = 900
#       service_account_name  = google_service_account.runner_service_account.email
#     }
#     metadata {
#       annotations = {
#         "autoscaling.knative.dev/maxScale" = "3"
#       }
#     }
#   }

#   metadata {
#     annotations = {
#       "run.googleapis.com/client-name" = "terraform"
#       "run.googleapis.com/ingress"     = "all"
#       # all, internal, internal-and-cloud-load-balancing
#     }
#   }

#   traffic {
#     percent         = 100
#     latest_revision = true
#   }

#   depends_on = [google_secret_manager_secret_version.config_secret_version]
# }


# resource "google_cloud_run_service_iam_member" "app-public-access" {
#   location = google_cloud_run_service.app.location
#   project  = google_cloud_run_service.app.project
#   service  = google_cloud_run_service.app.name
#   role     = "roles/run.invoker"
#   member   = "serviceAccount:${google_service_account.runner_service_account.email}"
# }