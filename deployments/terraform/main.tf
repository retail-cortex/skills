# Enterprise Infrastructure Provisioning for Retail Cortex Skill Builder

# 1. Enable Required Google Cloud APIs
resource "google_project_service" "apis" {
  for_each = toset([
    "compute.googleapis.com",
    "container.googleapis.com",
    "alloydb.googleapis.com",
    "servicenetworking.googleapis.com",
    "aiplatform.googleapis.com",
    "secretmanager.googleapis.com",
    "cloudtrace.googleapis.com",
    "logging.googleapis.com",
    "monitoring.googleapis.com",
    "artifactregistry.googleapis.com"
  ])

  project            = var.project_id
  service            = each.key
  disable_on_destroy = false
}

# 2. Network Module: VPC, Subnets, PSA Peering for AlloyDB, Cloud NAT
module "network" {
  source = "./modules/network"

  project_id   = var.project_id
  region       = var.region
  network_name = var.network_name

  depends_on = [google_project_service.apis]
}

# 3. AlloyDB Module: Cluster, Primary Instance, Databases (dev, qa, prod), Users, Secrets
module "alloydb" {
  source = "./modules/alloydb"

  project_id                        = var.project_id
  region                            = var.region
  network_id                        = module.network.network_id
  cluster_id                        = "skill-builder-alloydb"
  primary_instance_id               = "primary-instance"
  cpu_count                         = var.alloydb_cpu_count
  db_environments                   = var.environments
  private_vpc_connection_dependency = module.network.private_vpc_connection

  depends_on = [
    google_project_service.apis,
    module.network
  ]
}

# 4. GKE Module: 3 VPC-Native Private GKE Clusters (dev, qa, prod) with Workload Identity
module "gke" {
  source = "./modules/gke"

  project_id   = var.project_id
  region       = var.region
  network_name = module.network.network_name
  subnets      = module.network.subnets
  environments = var.gke_node_configs

  depends_on = [
    google_project_service.apis,
    module.network
  ]
}
