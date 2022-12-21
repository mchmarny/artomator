locals {
  kms_roles = toset([
    "roles/cloudkms.cryptoKeyEncrypterDecrypter",
    "roles/cloudkms.signerVerifier",
    "roles/cloudkms.viewer",
  ])
}

resource "google_kms_key_ring" "keyring" {
  name     = "${var.name}-signer-ring"
  location = "global"
}
resource "google_kms_crypto_key" "key" {
  name            = "${var.name}-signer-key"
  key_ring        = google_kms_key_ring.keyring.id
  lifecycle {
    prevent_destroy = true
  }
}

resource "google_kms_crypto_key_iam_member" "crypto_key_ops" {
  for_each      = local.kms_roles
  crypto_key_id = google_kms_crypto_key.key.id
  role          = each.value
  member        = "serviceAccount:${google_service_account.github_actions_user.email}"
}

