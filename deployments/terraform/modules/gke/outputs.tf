output "cluster_names" {
  description = "Map of GKE cluster names for dev, qa, and prod"
  value = {
    for k, v in google_container_cluster.clusters : k => v.name
  }
}

output "cluster_endpoints" {
  description = "Map of GKE cluster API master endpoints"
  value = {
    for k, v in google_container_cluster.clusters : k => v.endpoint
  }
}

output "cluster_ca_certificates" {
  description = "Map of GKE cluster CA certificates"
  value = {
    for k, v in google_container_cluster.clusters : k => v.master_auth[0].cluster_ca_certificate
  }
  sensitive = true
}

output "workload_identity_service_accounts" {
  description = "Google Service Account emails configured for Workload Identity"
  value = {
    for k, v in google_service_account.skill_service_sa : k => v.email
  }
}
