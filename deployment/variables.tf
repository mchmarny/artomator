# List of variables which can be provided ar runtime to override the specified defaults 

variable "project_id" {
  description = "GCP Project ID"
  type        = string
  nullable    = false
}

variable "name" {
  description = "Base name to derive everythign else from"
  default     = "artomator"
  type        = string
  nullable    = false
}

variable "location" {
  description = "Deployment location"
  default     = "us-west1"
  type        = string
  nullable    = false
}

variable "git_repo" {
  description = "GitHub Repo"
  type        = string
  nullable    = false
}

variable "image" {
  description = "Image URI"
  default     = "us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator"
  type        = string
  nullable    = false
}

variable "runtime_only" {
  description = "Whether or not deploy the development resoruces"
  default     = true
  type        = bool
}
