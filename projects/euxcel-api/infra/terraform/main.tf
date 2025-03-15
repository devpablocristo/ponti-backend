provider "vault" {
  address = var.vault_address
  token   = var.vault_token
}

provider "consul" {
  address = var.consul_address
}

resource "vault_generic_secret" "app_secrets" {
  path = "secret/staging/app"
  data_json = jsonencode({
    APP_NAME      = "qh"
    APP_VERSION   = "1.0"
    API_KEY       = "example-api-key"
    DATABASE_PASS = "example-db-pass"
  })
}

resource "consul_key_prefix" "app_config" {
  path_prefix = "staging/app/"
  subkeys = {
    "name"        = "qh"
    "version"     = "1.0"
    "api_version" = "v1"
    "env"         = "staging"
  }
}

provider "local" {}

resource "local_file" "example" {
  filename = "example.txt"
  content  = "Hello, Terraform!"
}

