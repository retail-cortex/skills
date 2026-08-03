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

# Service Account for ADK Agent Execution
resource "google_service_account" "agent_runner" {
  account_id   = "adk-agent-runner-sa"
  display_name = "ADK Agent Cloud Run Service Account"
}

# BigQuery Dataset
resource "google_bigquery_dataset" "analytics_dataset" {
  dataset_id                  = "retail_analytics"
  friendly_name               = "Retail Analytics Dataset"
  location                    = "US"
  default_table_expiration_ms = 3600000
}

# Cloud Run Service hosting FastAPI ADK agent
resource "google_cloud_run_v2_service" "agent_service" {
  name     = "adk-agent-api"
  location = var.region

  template {
    service_account = google_service_account.agent_runner.email
    containers {
      image = "gcr.io/${var.project_id}/adk-agent-api:latest"
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
