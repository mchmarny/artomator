resource "google_storage_bucket" "artifact_store" {
  name          = "${var.name}-${var.project_id}"
  location      = var.location
  storage_class = "STANDARD"
  force_destroy = true

  uniform_bucket_level_access = true
}

resource "google_storage_bucket_iam_binding" "artifact_store_binding" {
  bucket = google_storage_bucket.artifact_store.name
  role   = "roles/storage.admin"
  members = [
    "serviceAccount:${google_service_account.runner_service_account.email}",
  ]
}
