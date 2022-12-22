resource "google_kms_key_ring" "keyring" {
  name     = "${var.name}-signer-ring"
  location = "global"
}

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

resource "google_project_iam_member" "project" {
  project = var.project_id
  role    = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member  = "serviceAccount:${google_service_account.github_actions_user.email}"
}

# resource "google_service_account_iam_binding" "admin-account-iam" {
#   service_account_id = "serviceAccount:${google_service_account.github_actions_user.email}"
#   role               = "roles/cloudkms.cryptoKeyEncrypterDecrypter"

#   members = [
#     "user:${google_service_account.github_actions_user.email}",
#   ]
# }


resource "google_kms_crypto_key_iam_binding" "crypto_key_bindng" {
  crypto_key_id = google_kms_crypto_key.key.id
  role          = "roles/cloudkms.admin"
  members = [
    "serviceAccount:${google_service_account.github_actions_user.email}",
  ]
}

