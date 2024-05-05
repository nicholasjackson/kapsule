resource "container" "vault" {
  image {
    name = "hashicorp/vault:1.16.2"
  }

  network {
    id         = resource.network.local.meta.id
  }

  port {
    local = 8200
    host  = 8200
  }

  environment = {
    VAULT_DEV_ROOT_TOKEN_ID = "root"
    VAULT_DEV_LISTEN_ADDRESS = "0.0.0.0:8200"
  }

  health_check {
    timeout = "30s"
    http {
      address = "http://localhost:8200/v1/sys/health"
      success_codes = [200]
    }
  }
}

resource "exec" "vault" {
  script = file("./files/configure_vault.sh")
  target = resource.container.vault

  environment = {
    VAULT_ADDR = "http://localhost:8200"
    VAULT_TOKEN = "root"
  }
}

output "VAULT_ADDR" {
  value = "http://${resource.container.vault.container_name}:8200"
}

output "VAULT_TOKEN" {
  value = "root"
}