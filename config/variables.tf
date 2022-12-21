# List of variables which can be provided ar runtime to override the specified defaults 

variable "project_id" {
  description = "GCP Project ID"
  type        = string
  nullable    = false
}

variable "registry_location" {
  description = "Location of the Artifact Registry"
  type        = string
  nullable    = false
}

variable "registry_name" {
  description = "Name (ID) of the Artifact Registry"
  type        = string
  nullable    = false
}

variable "git_repo" {
  description = "GitHub Repo"
  type        = string
  nullable    = false
}
