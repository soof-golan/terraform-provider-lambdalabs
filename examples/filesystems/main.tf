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


data "lambdalabs_filesystems" "edu" {}

output "lambda_filesystems" {
  value = data.lambdalabs_filesystems.edu
}