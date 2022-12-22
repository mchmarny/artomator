# List of outputs from each terraform apply 

output "PROJECT_ID" {
  value       = data.google_project.project.name
  description = "Project ID to use in Auth action for GCP in GitHub."
}

output "CI_SERVICE_ACCOUNT" {
  count      = var.runtime_only ? 0 : 1
  value       = google_service_account.github_actions_user.email
  description = "Service account to use in GitHub Action for federated auth."
}

output "RUN_SERVICE_ACCOUNT" {
  value       = google_service_account.runner_service_account.email
  description = "Service account running Cloud Run service."
}

output "IDENTITY_PROVIDER" {
  count      = var.runtime_only ? 0 : 1
  value       = google_iam_workload_identity_pool_provider.github_provider.name
  description = "Provider ID to use in Auth action for GCP in GitHub."
}

output "KMS_KEY" {
  value       = google_kms_crypto_key.key.id
  description = "Cosign-formated URI to the signing key."
}

output "REGISTRY_URI" {
  count      = var.runtime_only ? 0 : 1
  value       = "${google_artifact_registry_repository.registry.location}-docker.pkg.dev/${data.google_project.project.name}/${google_artifact_registry_repository.registry.name}"
  description = "Artifact Registry location."
}

output "SERVING_IMAGE" {
  count      = var.runtime_only ? 0 : 1
  value       = "${var.image}:${data.template_file.version.rendered}"
  description = "Image currently being used in Cloud Run."
}

output "SERVICE_URL" {
  value       = google_cloud_run_service.app.status[0].url
  description = "Cloud Run service URL."
}

output "ARTIFACT_BUCKET" {
  value       = google_storage_bucket.artifact_store.url
  description = "GCS Bucket where exported artifacts will be saved"
}



