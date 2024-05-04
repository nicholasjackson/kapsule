resource "certificate_leaf" "registry" {
  ca_key  = "${jumppad()}/certs/root.key"
  ca_cert = "${jumppad()}/certs/root.cert"

  ip_addresses = ["127.0.0.1", variable.auth_ip_address]

  dns_names = [
    "localhost",
    "auth-registry.demo.gs",
    "auth.container.local.jmpd.in",
  ]

  output = data("certs")
}

resource "container" "auth" {
  image {
    name = "registry:2"
  }

  network {
    id         = resource.network.local.meta.id
    ip_address = variable.auth_ip_address
  }

  port {
    local = 443
    host  = 5001
  }

  environment = {
    DEBUG                         = "true"
    REGISTRY_HTTP_ADDR            = "0.0.0.0:443"
    REGISTRY_AUTH                 = "htpasswd"
    REGISTRY_AUTH_HTPASSWD_REALM  = "Registry Realm"
    REGISTRY_AUTH_HTPASSWD_PATH   = "/etc/auth/htpasswd"
    REGISTRY_HTTP_TLS_CERTIFICATE = "/certs/registry-leaf.cert"
    REGISTRY_HTTP_TLS_KEY         = "/certs/registry-leaf.key"
  }

  volume {
    source      = "./files/htpasswd"
    destination = "/etc/auth/htpasswd"
  }

  volume {
    source      = data("certs")
    destination = "/certs"
  }
}