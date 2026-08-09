output "vpc_network_name" {
  description = "Name of the provisioned VPC network"
  value       = module.network.network_name
}

output "vpc_network_id" {
  description = "ID of the provisioned VPC network"
  value       = module.network.network_id
}

output "alloydb_cluster_id" {
  description = "AlloyDB Cluster Identifier"
  value       = module.alloydb.cluster_id
}

output "alloydb_primary_ip" {
  description = "Private IP address of the AlloyDB Primary Instance"
  value       = module.alloydb.primary_instance_ip
}

output "alloydb_connection_secrets" {
  description = "Secret Manager secret names containing DSN connection strings for each environment"
  value       = module.alloydb.secret_manager_secrets
}

output "gke_clusters" {
  description = "Map of created GKE cluster names"
  value       = module.gke.cluster_names
}

output "gke_cluster_endpoints" {
  description = "Map of GKE cluster master endpoints"
  value       = module.gke.cluster_endpoints
}

output "workload_identity_service_accounts" {
  description = "Service accounts provisioned for GKE Workload Identity"
  value       = module.gke.workload_identity_service_accounts
}
