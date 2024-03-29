# Description: Creates a KMS key ring and key for signing artifacts

# Creates a KMS key ring and key for signing artifacts
resource "google_kms_key_ring" "keyring" {
  name     = "${var.name}-signer-ring"
  location = "global"
}

# Creates a KMS key for signing artifacts
resource "google_kms_crypto_key" "key" {
  name     = "${var.name}-signer"
  key_ring = google_kms_key_ring.keyring.id
  purpose  = "ASYMMETRIC_SIGN"

  version_template {
    algorithm = "RSA_SIGN_PKCS1_4096_SHA512"
  }

  lifecycle {
    prevent_destroy = true
  }
}

data "google_kms_crypto_key_version" "version" {
  crypto_key = google_kms_crypto_key.key.id
}

# Binds the runner service account to the key with key encrypter/decrypter permissions
resource "google_kms_crypto_key_iam_binding" "crypto_signer_verifier" {
  count         = var.runtime_only ? 0 : 1
  crypto_key_id = google_kms_crypto_key.key.id
  role          = "roles/cloudkms.signerVerifier"
  members = [
    "serviceAccount:${google_service_account.github_actions_user[count.index].email}",
  ]
}

# Binds the runner service account to the key with key viewer permissions
resource "google_kms_crypto_key_iam_binding" "crypto_viewer" {
  count         = var.runtime_only ? 0 : 1
  crypto_key_id = google_kms_crypto_key.key.id
  role          = "roles/cloudkms.viewer"
  members = [
    "serviceAccount:${google_service_account.github_actions_user[count.index].email}",
  ]
}


# Binds the runner service account to the key with key operator permissions
resource "google_kms_crypto_key_iam_binding" "crypto_crypto_Operator" {
  count         = var.runtime_only ? 0 : 1
  crypto_key_id = google_kms_crypto_key.key.id
  role          = "roles/cloudkms.cryptoOperator"
  members = [
    "serviceAccount:${google_service_account.github_actions_user[count.index].email}",
  ]
}
