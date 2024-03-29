# Description: Outputs for the deployment

output "PROJECT_ID" {
  value       = data.google_project.project.name
  description = "Project ID to use in Auth action for GCP in GitHub."
}

output "CI_SERVICE_ACCOUNT" {
  value       = var.runtime_only ? null : google_service_account.github_actions_user[0].email
  description = "Service account to use in GitHub Action for federated auth."
}

output "RUN_SERVICE_ACCOUNT" {
  value       = google_service_account.runner_service_account.email
  description = "Service account running Cloud Run service."
}

output "IDENTITY_PROVIDER" {
  value       = var.runtime_only ? null : google_iam_workload_identity_pool_provider.github_provider[0].name
  description = "Provider ID to use in Auth action for GCP in GitHub."
}

output "KMS_KEY" {
  value       = data.google_kms_crypto_key_version.version.name
  description = "Cosign-formated URI to the signing key."
}

output "REGISTRY_URI" {
  value       = var.runtime_only ? null : "${google_artifact_registry_repository.registry[0].location}-docker.pkg.dev"
  description = "Artifact Registry location."
}

output "SERVING_IMAGE" {
  value       = var.runtime_only ? null : "${var.image}:${data.template_file.version.rendered}"
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

output "ATTESTOR_ID" {
  value       = google_binary_authorization_attestor.attestor.id
  description = "Binary Authorization attestor name."
}
