
resource "google_artifact_registry_repository" "registry" {
  provider = google-beta
  project = var.project_id
  description = "${var.name} artifacts registry"
  location = var.location
  repository_id = var.name
  format = "DOCKER"
}