variable "lambdalabs_api_key" {
  description = "Lambda Labs API Key"
}

provider "lambdalabs" {
  api_key = var.lambdalabs_api_key
}
