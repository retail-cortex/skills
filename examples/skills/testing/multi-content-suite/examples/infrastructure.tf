terraform {
  required_version = ">= 1.5.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

resource "google_storage_bucket" "asset_store" {
  name          = "retail-cortex-multi-content-assets"
  location      = "US"
  force_destroy = false
  uniform_bucket_level_access = true
}
