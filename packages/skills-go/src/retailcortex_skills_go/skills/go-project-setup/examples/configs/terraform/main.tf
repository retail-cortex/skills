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
  description = "GCP Compute Region"
}

# GKE Enterprise Cluster for Go Microservices
resource "google_container_cluster" "primary" {
  name     = "go-enterprise-gke"
  location = var.region
  enable_autopilot = true
}
