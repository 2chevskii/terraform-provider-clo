terraform {
  required_providers {
    clo = {
      source  = "2chevskii/clo"
      version = "2.8.1-dev.0"
    }
  }
}

provider "clo" {
  auth_url = "https://api.clo.ru"
  # Set CLO_API_AUTH_TOKEN in the process environment.
}

data "clo_projects" "available" {}
