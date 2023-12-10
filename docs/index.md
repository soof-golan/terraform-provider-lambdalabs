---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "lambdalabs Provider"
subcategory: ""
description: |-
  
---

# lambdalabs Provider



## Example Usage

```terraform
variable "lambdalabs_api_key" {
  description = "Lambda Labs API Key"
}

provider "lambdalabs" {
  api_key = var.lambdalabs_api_key
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String, Sensitive) Lambda Labs API key
- `host` (String) Lambda Labs API host