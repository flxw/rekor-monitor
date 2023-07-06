resource "digitalocean_vpc" "default-fra1" {
  name   = "default-fra1"
  region = "fra1"
}

resource "digitalocean_database_cluster" "postgres-cluster" {
  name       = "app-acd8ae2b-31d1-4da7-b4d5-2203ba9a3f8d"
  engine     = "pg"
  node_count = 1
  region     = "fra1"
  size       = "db-s-1vcpu-1gb"
  version    = 12
  private_network_uuid = digitalocean_vpc.default-fra1.id
}

resource "digitalocean_database_db" "monitor-db" {
  cluster_id = digitalocean_database_cluster.postgres-cluster.id
  name       = "rekord-monitor-dev"
}

resource "digitalocean_database_user" "monitor-db-user" {
  cluster_id = digitalocean_database_cluster.postgres-cluster.id
  name       = "rekord-monitor-dev"
}

resource "digitalocean_app" "rekor-monitor" {
  spec {
    name   = "rekor-monitor"
    region = "fra"

    /*service {
      name           = "grafana"
      instance_count = 1
      http_port      = 3000
      image {
        registry_type = "DOCKER_HUB"
        registry      = "grafana"
        repository    = "grafana-enterprise"
      }
    }*/

    database {
      cluster_name = digitalocean_database_cluster.postgres-cluster.name
      db_name      = "rekord-monitor-dev"
      db_user      = "rekord-monitor-dev"
      engine       = "PG"
      name         = "rekord-monitor-dev"
      production   = true
    }

    worker {
      instance_size_slug = "basic-xxs"
      name               = "rekor-crawler"
      instance_count     = 1
      dockerfile_path    = "Dockerfile"

      github {
        repo   = "flxw/rekor-monitor"
        branch = "master"
      }

      env {
        key   = "REKOR_START_INDEX"
        value = 24686621
        type  = "GENERAL"
      }

      env {
        key   = "DATABASE_URL"
        value = "postgres://${digitalocean_database_user.monitor-db-user.name}:${digitalocean_database_user.monitor-db-user.password}@${digitalocean_database_cluster.postgres-cluster.host}:${digitalocean_database_cluster.postgres-cluster.port}/${digitalocean_database_db.monitor-db.name}"
        type  = "SECRET"
      }
    }
  }
}
