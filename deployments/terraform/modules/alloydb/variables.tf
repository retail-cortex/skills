variable "project_id" {
  description = "Google Cloud Project ID"
  type        = string
}

variable "region" {
  description = "Google Cloud Region"
  type        = string
  default     = "us-central1"
}

variable "network_id" {
  description = "The VPC network self link or ID"
  type        = string
}

variable "cluster_id" {
  description = "ID of the AlloyDB cluster"
  type        = string
  default     = "skill-builder-alloydb-cluster"
}

variable "primary_instance_id" {
  description = "ID of the primary AlloyDB instance"
  type        = string
  default     = "primary-instance"
}

variable "cpu_count" {
  description = "Number of vCPUs for the primary instance (2, 4, 8, 16, etc.)"
  type        = number
  default     = 2
}

variable "db_environments" {
  description = "Environments requiring dedicated databases and users"
  type        = list(string)
  default     = ["dev", "qa", "prod"]
}

variable "private_vpc_connection_dependency" {
  description = "Explicit dependency ensuring Service Networking peering is ready"
  type        = any
  default     = null
}
