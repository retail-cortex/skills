terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.30.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.30.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6.0"
    }
  }

  # Production remote state storage in Google Cloud Storage
  # Configure via: terraform init -backend-config="bucket=<YOUR_TF_STATE_BUCKET>" -backend-config="prefix=skill-builder"
  # backend "gcs" {
  #   bucket = "YOUR_GCS_TERRAFORM_STATE_BUCKET"
  #   prefix = "skill-builder/deployments"
  # }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}
