terraform {
  required_version = ">= 1.5.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0.0"
    }
  }
}

# 1. VPC Network
resource "google_compute_network" "vpc" {
  name                    = var.network_name
  project                 = var.project_id
  auto_create_subnetworks = false
  routing_mode            = "REGIONAL"
  description             = "Dedicated VPC for Enterprise Skill Builder GKE clusters and AlloyDB AI"
}

# 2. Subnets for GKE Environments (VPC-Native with Alias IP Ranges)
resource "google_compute_subnetwork" "subnets" {
  for_each = var.subnets

  name                     = "${var.network_name}-${each.key}-subnet"
  project                  = var.project_id
  region                   = var.region
  network                  = google_compute_network.vpc.id
  ip_cidr_range            = each.value.ip_cidr_range
  private_ip_google_access = true

  secondary_ip_range {
    range_name    = "${each.key}-pods"
    ip_cidr_range = each.value.pod_cidr
  }

  secondary_ip_range {
    range_name    = "${each.key}-services"
    ip_cidr_range = each.value.svc_cidr
  }

  log_config {
    aggregation_interval = "INTERVAL_5_SEC"
    flow_sampling        = 0.5
    metadata             = "INCLUDE_ALL_METADATA"
  }
}

# 3. Private Services Access for AlloyDB AI / Cloud SQL (VPC Peering)
resource "google_compute_global_address" "private_ip_alloc" {
  name          = "${var.network_name}-alloydb-range"
  project       = var.project_id
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.vpc.id
  description   = "Reserved internal IP range for AlloyDB and Service Networking peering"
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_alloc.name]
}

# 4. Cloud Router & Cloud NAT for Private GKE Node Internet Access (Egress)
resource "google_compute_router" "router" {
  name    = "${var.network_name}-router"
  project = var.project_id
  region  = var.region
  network = google_compute_network.vpc.id
}

resource "google_compute_router_nat" "nat" {
  name                               = "${var.network_name}-nat"
  project                            = var.project_id
  router                             = google_compute_router.router.name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}

# 5. Firewall Rules
resource "google_compute_firewall" "allow_internal" {
  name        = "${var.network_name}-allow-internal"
  project     = var.project_id
  network     = google_compute_network.vpc.name
  description = "Allow internal traffic within the VPC network"

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }

  allow {
    protocol = "udp"
    ports    = ["0-65535"]
  }

  source_ranges = [
    "10.0.0.0/8",
    "172.16.0.0/12",
    "192.168.0.0/16"
  ]
}

resource "google_compute_firewall" "allow_health_checks" {
  name        = "${var.network_name}-allow-health-checks"
  project     = var.project_id
  network     = google_compute_network.vpc.name
  description = "Allow Google Cloud Load Balancer health checks"

  allow {
    protocol = "tcp"
    ports    = ["80", "443", "8000", "8080", "9090"]
  }

  source_ranges = [
    "35.191.0.0/16",
    "130.211.0.0/22"
  ]
}
