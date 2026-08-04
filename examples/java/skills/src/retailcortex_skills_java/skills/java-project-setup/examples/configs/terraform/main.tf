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
  description = "Compute region"
}

resource "google_cloud_run_v2_service" "java_service" {
  name     = "enterprise-java-service"
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project_id}/enterprise-java-service:v1.0.0"
      ports {
        container_port = 8080
      }
    }
  }
}
