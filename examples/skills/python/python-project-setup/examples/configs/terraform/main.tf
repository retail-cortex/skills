terraform {
  required_version = ">= 1.9.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

variable "project_id" {
  type        = string
  description = "GCP Project ID"
}

variable "region" {
  type        = string
  default     = "us-central1"
  description = "Primary compute region"
}

resource "google_cloud_run_v2_service" "python_service" {
  name     = "enterprise-python-service"
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project_id}/enterprise-python-service:v1.0.0"
      ports {
        container_port = 8000
      }
      env {
        name  = "GOOGLE_CLOUD_PROJECT"
        value = var.project_id
      }
    }
  }
}
