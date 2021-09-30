provider "google" {
  project = var.project
  region  = var.region
}

resource "google_cloud_run_service" "default" {
  name     = var.service_name
  location = var.region

  template {
    spec {
      containers {
        image = var.container_image
        dynamic "env" {
          for_each = var.config
          content {
            name  = env.value["name"]
            value = env.value["value"]
          }
        }
      }
    }
  }
}

resource "google_cloud_run_domain_mapping" "default" {
  count    = var.dns_domain != "" ? 1 : 0
  location = var.region
  name     = var.dns_domain

  metadata {
    namespace = var.project
  }

  spec {
    route_name = google_cloud_run_service.default.name
  }
}

data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "noauth" {
  location = google_cloud_run_service.default.location
  project  = google_cloud_run_service.default.project
  service  = google_cloud_run_service.default.name

  policy_data = data.google_iam_policy.noauth.policy_data
}
