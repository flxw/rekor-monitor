resource "digitalocean_vpc" "default-fra1" {
  name   = "default-fra1"
  region = "fra1"
}

resource "digitalocean_database_cluster" "postgres-cluster" {
  name                 = "app-acd8ae2b-31d1-4da7-b4d5-2203ba9a3f8d"
  engine               = "pg"
  node_count           = 1
  region               = "fra1"
  size                 = "db-s-1vcpu-1gb"
  version              = 12
  private_network_uuid = digitalocean_vpc.default-fra1.id
}

resource "digitalocean_database_db" "monitor-db" {
  cluster_id = digitalocean_database_cluster.postgres-cluster.id
  name       = "rekord-monitor-dev"
}

resource "digitalocean_database_db" "grafana-db" {
  cluster_id = digitalocean_database_cluster.postgres-cluster.id
  name       = "grafana"
}

resource "digitalocean_database_user" "monitor-db-user" {
  cluster_id = digitalocean_database_cluster.postgres-cluster.id
  name       = "rekord-monitor-dev"
}

resource "digitalocean_database_user" "grafana" {
  cluster_id = digitalocean_database_cluster.postgres-cluster.id
  name       = "grafana-read"
}

resource "digitalocean_database_user" "grafana-db-user" {
  cluster_id = digitalocean_database_cluster.postgres-cluster.id
  name       = "grafana-db-user"
}

resource "digitalocean_app" "rekor-monitor" {
  spec {
    name   = "rekor-monitor"
    region = "fra"

    service {
      name               = "grafana"
      instance_count     = 1
      instance_size_slug = "basic-xxs"
      http_port          = 3000
      dockerfile_path    = "Dockerfile.grafana"
      source_dir         = "/"

      github {
        repo           = "flxw/rekor-monitor"
        branch         = "master"
        deploy_on_push = true
      }

      env {
        key   = "GF_SECURITY_ADMIN_USER"
        value = var.admin_user_name
      }

      env {
        key   = "GF_SECURITY_ADMIN_PASSWORD"
        value = var.admin_user_password
      }

      env {
        key   = "GF_FEATURE_TOGGLES_ENABLE"
        value = "publicDashboards"
      }

      env {
        key = "GF_DATABASE_TYPE"
        value = "postgres"
      }

      env {
        key = "GF_DATABASE_URL"
        value = "postgres://${digitalocean_database_user.grafana-db-user.name}:${digitalocean_database_user.grafana-db-user.password}@${digitalocean_database_cluster.postgres-cluster.host}:${digitalocean_database_cluster.postgres-cluster.port}/${digitalocean_database_db.grafana-db.name}"
        type = "SECRET"
      }

      env {
        key = "GF_DATABASE_SSL_MODE"
        value = "require"
      }

      env {
        key   = "DB_URL"
        value = "${digitalocean_database_cluster.postgres-cluster.host}:${digitalocean_database_cluster.postgres-cluster.port}"
        type = "SECRET"
      }

      env {
        key   = "DB_USER_NAME"
        value = digitalocean_database_user.monitor-db-user.name
      }

      env {
        key   = "DB_USER_PASSWORD"
        value = digitalocean_database_user.monitor-db-user.password
      }

      env {
        key   = "DB_NAME"
        value = digitalocean_database_db.monitor-db.name
      }

      env {
        key   = "DB_SSL_MODE"
        value = "require"
      }
    }

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
        repo           = "flxw/rekor-monitor"
        branch         = "master"
        deploy_on_push = false
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
