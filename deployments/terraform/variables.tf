variable "project_id" {
  description = "The Google Cloud Project ID"
  type        = string
}

variable "region" {
  description = "The primary Google Cloud region for infrastructure deployment"
  type        = string
  default     = "us-central1"
}

variable "network_name" {
  description = "Name of the custom VPC network"
  type        = string
  default     = "skill-builder-vpc"
}

variable "alloydb_cpu_count" {
  description = "Number of vCPUs for the primary AlloyDB AI instance (2, 4, 8, etc.)"
  type        = number
  default     = 2
}

variable "environments" {
  description = "Environments to deploy (dev, qa, prod)"
  type        = list(string)
  default     = ["dev", "qa", "prod"]
}

variable "gke_node_configs" {
  description = "Node pool sizing per environment"
  type = map(object({
    min_nodes    = number
    max_nodes    = number
    machine_type = string
    disk_size_gb = number
    preemptible  = bool
  }))
  default = {
    dev = {
      min_nodes    = 1
      max_nodes    = 3
      machine_type = "e2-standard-4"
      disk_size_gb = 50
      preemptible  = true
    }
    qa = {
      min_nodes    = 1
      max_nodes    = 4
      machine_type = "e2-standard-4"
      disk_size_gb = 50
      preemptible  = false
    }
    prod = {
      min_nodes    = 3
      max_nodes    = 10
      machine_type = "n2-standard-4"
      disk_size_gb = 100
      preemptible  = false
    }
  }
}
