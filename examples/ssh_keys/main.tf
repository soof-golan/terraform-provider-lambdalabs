terraform {
  required_providers {
    lambdalabs = {
      source = "hashicorp.com/edu/lambdalabs"
    }
  }
}

variable "lambdalabs_api_key" {
  description = "Lambda Labs API Key"
}

provider "lambdalabs" {
  api_key = var.lambdalabs_api_key
}

# Provision a new SSH Key
resource "lambdalabs_ssh_key" "edu" {
  name       = "example-key"
  public_key = trimspace(file("~/.ssh/id_ed25519.pub"))
}

# List Just One SSH Key by Name
data "lambdalabs_ssh_key" "edu" {
  depends_on = [lambdalabs_ssh_key.edu]
  name       = lambdalabs_ssh_key.edu.name
}

# List ALl SSH Keys
data "lambdalabs_ssh_keys" "edu" {
  depends_on = [lambdalabs_ssh_key.edu]
}

output "lambda_ssh_keys" {
  value = data.lambdalabs_ssh_keys.edu
}

output "lambda_ssh_key" {
  value = data.lambdalabs_ssh_key.edu
}

output "lambda_ssh_key_resource" {
  value = lambdalabs_ssh_key.edu
}
