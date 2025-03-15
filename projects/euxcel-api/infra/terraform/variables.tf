variable "vault_address" {
  description = "The address of the Vault server"
  type        = string
  default     = "http://localhost:8200"
}

variable "vault_token" {
  description = "The Vault root token"
  type        = string
  default     = "root"
}

variable "consul_address" {
  description = "The address of the Consul server"
  type        = string
  default     = "http://localhost:8500"
}
