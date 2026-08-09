terraform {
  required_version = ">= 1.5.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5.0"
    }
  }
}

# 1. Generate Root Superuser Password
resource "random_password" "alloydb_root_password" {
  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

# 2. AlloyDB AI Cluster Definition
resource "google_alloydb_cluster" "cluster" {
  cluster_id   = var.cluster_id
  location     = var.region
  project      = var.project_id
  cluster_type = "PRIMARY"

  network_config {
    network = var.network_id
  }

  initial_user {
    user     = "postgres"
    password = random_password.alloydb_root_password.result
  }

  automated_backup_policy {
    enabled = true
    weekly_schedule {
      days_of_week = ["MONDAY", "WEDNESDAY", "FRIDAY", "SUNDAY"]
      start_times {
        hours   = 2
        minutes = 0
      }
    }
    backup_window = "3600s"
    quantity_based_retention {
      count = 14
    }
  }

  continuous_backup_config {
    enabled              = true
    recovery_window_days = 7
  }

  labels = {
    managed_by  = "terraform"
    application = "skill-builder"
  }

  depends_on = [var.private_vpc_connection_dependency]
}

# 3. Primary Instance with pgvector and Google ML Support
resource "google_alloydb_instance" "primary" {
  cluster       = google_alloydb_cluster.cluster.name
  instance_id   = var.primary_instance_id
  instance_type = "PRIMARY"

  machine_config {
    cpu_count = var.cpu_count
  }

  database_flags = {
    "google_ml.enable_model_support" = "on"
    "alloydb.enable_pgvector"        = "on"
    "password_encryption"            = "scram-sha-256"
    "max_connections"                = "500"
  }

  availability_type = "REGIONAL" # High availability across zones
}

# 4. Generate Passwords for Environment Users (dev, qa, prod)
resource "random_password" "env_user_passwords" {
  for_each         = toset(var.db_environments)
  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

# 5. Provision Dedicated Users for dev, qa, and prod
resource "google_alloydb_user" "env_users" {
  for_each       = toset(var.db_environments)
  cluster        = google_alloydb_cluster.cluster.name
  user_id        = "${each.key}_user"
  user_type      = "ALLOYDB_BUILT_IN"
  password       = random_password.env_user_passwords[each.key].result
  database_roles = ["alloydbsuperuser"]

  depends_on = [google_alloydb_instance.primary]
}

# 6. Store Database Credentials in Google Cloud Secret Manager
resource "google_secret_manager_secret" "db_connection_strings" {
  for_each  = toset(var.db_environments)
  secret_id = "skill-builder-alloydb-${each.key}-dsn"
  project   = var.project_id

  replication {
    auto {}
  }

  labels = {
    environment = each.key
    application = "skill-builder"
  }
}

resource "google_secret_manager_secret_version" "db_connection_strings" {
  for_each = toset(var.db_environments)
  secret   = google_secret_manager_secret.db_connection_strings[each.key].id
  secret_data = format(
    "host=%s port=5432 user=%s_user password=%s dbname=skills_%s sslmode=require",
    google_alloydb_instance.primary.ip_address,
    each.key,
    random_password.env_user_passwords[each.key].result,
    each.key
  )
}
