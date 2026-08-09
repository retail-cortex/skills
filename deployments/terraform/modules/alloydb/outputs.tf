output "cluster_id" {
  description = "The ID of the AlloyDB cluster"
  value       = google_alloydb_cluster.cluster.cluster_id
}

output "cluster_name" {
  description = "The full resource name of the AlloyDB cluster"
  value       = google_alloydb_cluster.cluster.name
}

output "primary_instance_ip" {
  description = "Private IP address of the primary AlloyDB instance"
  value       = google_alloydb_instance.primary.ip_address
}

output "database_users" {
  description = "Map of created database users"
  value = {
    for k, v in google_alloydb_user.env_users : k => v.user_id
  }
}

output "secret_manager_secrets" {
  description = "Secret Manager secret IDs for each environment database connection string"
  value = {
    for k, v in google_secret_manager_secret.db_connection_strings : k => v.secret_id
  }
}

output "database_names" {
  description = "Database names partitioned by environment"
  value = {
    for env in var.db_environments : env => "skills_${env}"
  }
}
