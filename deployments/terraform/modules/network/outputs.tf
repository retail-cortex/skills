output "network_id" {
  description = "The ID of the VPC network"
  value       = google_compute_network.vpc.id
}

output "network_name" {
  description = "The name of the VPC network"
  value       = google_compute_network.vpc.name
}

output "network_self_link" {
  description = "The self link of the VPC network"
  value       = google_compute_network.vpc.self_link
}

output "subnets" {
  description = "Map of subnets created with self links and secondary ranges"
  value = {
    for k, v in google_compute_subnetwork.subnets : k => {
      id                  = v.id
      name                = v.name
      self_link           = v.self_link
      ip_cidr_range       = v.ip_cidr_range
      secondary_pod_range = "${k}-pods"
      secondary_svc_range = "${k}-services"
    }
  }
}

output "private_vpc_connection" {
  description = "The private service connection for AlloyDB"
  value       = google_service_networking_connection.private_vpc_connection.id
}
