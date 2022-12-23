# Description: Redis instance and VPC connector

# Redis instance
resource "google_redis_instance" "cache" {
  name           = var.name
  region         = var.location
  redis_version  = "REDIS_5_0"
  tier           = "STANDARD_HA"
  memory_size_gb = 5
}

# VPC connector
resource "google_vpc_access_connector" "connector" {
  name          = "${var.name}-connector"
  region        = var.location
  ip_cidr_range = "10.10.0.0/28"
  network       = "default"
}