# INWX Provider

The INWX Provider can be used to register and manage domains and their domain contacts. 

## Example Usage

**Terraform 0.13+**

```terraform
terraform {
  required_providers {
    inwx = {
      source = "inwx/inwx"
      version = ">= 1.0.0"
    }
  }
}

// API configuration
provider "inwx" {
  api_url = "https://api.ote.domrobot.com/jsonrpc/"
  username = "example-user"
  password = "redacted"
  tan = "000000"
}

// contact used for domain
resource "inwx_domain_contact" "example_person" {
  // contact configuration
}

resource "inwx_domain" "example_com" {
  // domain configuration
  // ...
  contacts {
    // references to terraform managed contact "example_person"
    registrant = inwx_domain_contact.example_person.id
    admin  = inwx_domain_contact.example_person.id
    tech  = inwx_domain_contact.example_person.id
    billing  = inwx_domain_contact.example_person.id
  }
  // ...
}

resource "inwx_nameserver_record" "example_com" {
  // nameserver record configuration
}
```

## Argument Reference

* `api_url` - (Optional) URL of the RPC API endpoint. Use `https://api.domrobot.com/jsonrpc/` for production and `https://api.ote.domrobot.com/jsonrpc/` for testing. Default: `https://api.domrobot.com/jsonrpc/`
* `username` - (Required) Login username of the api
* `password` - (Required) Login password of the api
* `tan` - (Optional) [mobile tan](https://www.inwx.com/en/offer/mobiletan)