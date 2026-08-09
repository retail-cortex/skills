variable "project_id" {
  description = "The Google Cloud Project ID"
  type        = string
}

variable "region" {
  description = "The primary Google Cloud region"
  type        = string
  default     = "us-central1"
}

variable "network_name" {
  description = "Name of the VPC network"
  type        = string
  default     = "skill-builder-vpc"
}

variable "subnets" {
  description = "Subnet configurations for dev, qa, and prod environments"
  type = map(object({
    ip_cidr_range = string
    pod_cidr      = string
    svc_cidr      = string
  }))
  default = {
    dev = {
      ip_cidr_range = "10.10.0.0/20"
      pod_cidr      = "10.100.0.0/16"
      svc_cidr      = "10.101.0.0/20"
    }
    qa = {
      ip_cidr_range = "10.20.0.0/20"
      pod_cidr      = "10.102.0.0/16"
      svc_cidr      = "10.103.0.0/20"
    }
    prod = {
      ip_cidr_range = "10.30.0.0/20"
      pod_cidr      = "10.104.0.0/16"
      svc_cidr      = "10.105.0.0/20"
    }
  }
}
