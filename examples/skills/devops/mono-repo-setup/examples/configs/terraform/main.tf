terraform {
  required_version = ">= 1.9.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
  backend "gcs" {
    bucket = "enterprise-monorepo-tfstate"
    prefix = "monorepo/global"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

variable "project_id" {
  type        = string
  description = "Google Cloud Project ID"
}

variable "region" {
  type        = string
  default     = "us-central1"
  description = "GCP Compute Region"
}

# Artifact Registry for all polyglot container builds
resource "google_artifact_registry_repository" "monorepo_containers" {
  location      = var.region
  repository_id = "monorepo-containers"
  format        = "DOCKER"
  description   = "Centralized container registry for monorepo microservices"
}
