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

resource "random_string" "ssh_key_name" {
  length = 8
}

# Provision a new SSH Key
resource "lambdalabs_ssh_key" "instance_ssh_key" {
  name       = "instance-key-${random_string.ssh_key_name.result}"
  public_key = trimspace(file("~/.ssh/id_ed25519.pub"))
}

resource "lambdalabs_instance" "example_instance" {
  instance_type = "gpu_1x_a10" // This instance type costs about 60 cents an hour
  region        = "us-west-1"
  ssh_key_names = [lambdalabs_ssh_key.instance_ssh_key.name]
  filesystem_names = ["stable-diffusion"]
}

data "lambdalabs_instance" "example" {
  # Changes this to the instance id you want to query
  id = lambdalabs_instance.example_instance.id
}


data "lambdalabs_instances" "all_instances" {
  depends_on = [lambdalabs_instance.example_instance]
}

output "lambdalabs_instance_result" {
  value = lambdalabs_instance.example_instance
}


output "lambda_all_instances_result" {
  value = data.lambdalabs_instances.all_instances
}

output "lambda_just_one_instance_result" {
  value = data.lambdalabs_instance.example
}


