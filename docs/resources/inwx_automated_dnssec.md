# Resource: inwx_automated_dnssec

Automated DNSSEC management for a domain.

## Example Usage

```terraform
resource "inwx_automated_dnssec" "example_com" {
  domain = "example.com"
}
```

## Argument Reference

* `domain` - (Required) Name of the domain
