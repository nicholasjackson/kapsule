resource "network" "local" {
  subnet = "10.6.0.0/16"
}

variable "auth_ip_address" {
  default = "10.6.0.183"
}
