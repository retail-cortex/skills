terraform {
  required_version = ">= 1.5.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0.0"
    }
  }
}

# 1. Dedicated Service Account for GKE Nodes
resource "google_service_account" "gke_nodes" {
  for_each     = var.environments
  account_id   = "gke-node-sa-${each.key}"
  display_name = "GKE Nodes Service Account (${each.key})"
  project      = var.project_id
}

resource "google_project_iam_member" "gke_node_roles" {
  for_each = {
    for pair in setproduct(
      keys(var.environments),
      [
        "roles/logging.logWriter",
        "roles/monitoring.metricWriter",
        "roles/monitoring.viewer",
        "roles/stackdriver.resourceMetadata.writer",
        "roles/artifactregistry.reader"
      ]
      ) : "${pair[0]}-${pair[1]}" => {
      env  = pair[0]
      role = pair[1]
    }
  }

  project = var.project_id
  role    = each.value.role
  member  = "serviceAccount:${google_service_account.gke_nodes[each.value.env].email}"
}

# 2. GKE Clusters (dev, qa, prod)
resource "google_container_cluster" "clusters" {
  for_each = var.environments

  name     = "skill-builder-gke-${each.key}"
  project  = var.project_id
  location = var.region

  network    = var.network_name
  subnetwork = var.subnets[each.key].name

  # Remove default node pool immediately to manage custom node pool
  remove_default_node_pool = true
  initial_node_count       = 1

  # VPC-Native Cluster (Alias IPs)
  ip_allocation_policy {
    cluster_secondary_range_name  = var.subnets[each.key].secondary_pod_range
    services_secondary_range_name = var.subnets[each.key].secondary_svc_range
  }

  # Workload Identity Configuration
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  # Private Cluster Configuration
  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = cidrsubnet("172.16.0.0/16", 4, index(keys(var.environments), each.key))
  }

  release_channel {
    channel = each.key == "prod" ? "REGULAR" : "RAPID"
  }

  # Maintenance Policy for Prod
  dynamic "maintenance_policy" {
    for_each = each.key == "prod" ? [1] : []
    content {
      recurring_window {
        start_time = "2026-01-01T04:00:00Z"
        end_time   = "2026-01-01T08:00:00Z"
        recurrence = "FREQ=WEEKLY;BYDAY=SA,SU"
      }
    }
  }

  addons_config {
    http_load_balancing {
      disabled = false
    }
    horizontal_pod_autoscaling {
      disabled = false
    }
    network_policy_config {
      disabled = false
    }
    gcs_fuse_csi_driver_config {
      enabled = true
    }
  }

  network_policy {
    enabled  = true
    provider = "CALICO"
  }

  lifecycle {
    ignore_changes = [initial_node_count]
  }
}

# 3. Dedicated Autoscaling Node Pools
resource "google_container_node_pool" "node_pools" {
  for_each = var.environments

  name       = "primary-pool"
  project    = var.project_id
  location   = var.region
  cluster    = google_container_cluster.clusters[each.key].name
  node_count = each.value.min_nodes

  autoscaling {
    min_node_count = each.value.min_nodes
    max_node_count = each.value.max_nodes
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  node_config {
    machine_type = each.value.machine_type
    disk_size_gb = each.value.disk_size_gb
    disk_type    = "pd-balanced"
    preemptible  = each.value.preemptible

    service_account = google_service_account.gke_nodes[each.key].email
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    shielded_instance_config {
      enable_secure_boot          = true
      enable_integrity_monitoring = true
    }

    labels = {
      environment = each.key
      node_role   = "primary-worker"
      managed_by  = "terraform"
    }

    tags = ["skill-builder-gke", "${each.key}-node"]
  }
}

# 4. Google Service Accounts for Skill Service Workload Identity
resource "google_service_account" "skill_service_sa" {
  for_each     = var.environments
  account_id   = "skill-service-${each.key}-sa"
  display_name = "Skill Service Workload Identity SA (${each.key})"
  project      = var.project_id
}

# 5. IAM Bindings for Skill Service (Vertex AI, Cloud Trace, Secret Accessor)
resource "google_project_iam_member" "skill_service_vertex_ai" {
  for_each = var.environments
  project  = var.project_id
  role     = "roles/aiplatform.user"
  member   = "serviceAccount:${google_service_account.skill_service_sa[each.key].email}"
}

resource "google_project_iam_member" "skill_service_trace" {
  for_each = var.environments
  project  = var.project_id
  role     = "roles/cloudtrace.agent"
  member   = "serviceAccount:${google_service_account.skill_service_sa[each.key].email}"
}

resource "google_project_iam_member" "skill_service_secrets" {
  for_each = var.environments
  project  = var.project_id
  role     = "roles/secretmanager.secretAccessor"
  member   = "serviceAccount:${google_service_account.skill_service_sa[each.key].email}"
}

# 6. Workload Identity IAM User Binding to Kubernetes ServiceAccount
resource "google_service_account_iam_member" "workload_identity_binding" {
  for_each           = var.environments
  service_account_id = google_service_account.skill_service_sa[each.key].name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[skill-builder/skill-service-sa]"
}
