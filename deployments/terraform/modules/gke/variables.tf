variable "project_id" {
  description = "Google Cloud Project ID"
  type        = string
}

variable "region" {
  description = "Google Cloud Region"
  type        = string
  default     = "us-central1"
}

variable "network_name" {
  description = "The VPC network name"
  type        = string
}

variable "subnets" {
  description = "Map of subnets from the network module"
  type = map(object({
    id                  = string
    name                = string
    self_link           = string
    ip_cidr_range       = string
    secondary_pod_range = string
    secondary_svc_range = string
  }))
}

variable "environments" {
  description = "Configuration for each GKE cluster environment (dev, qa, prod)"
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
