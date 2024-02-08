# Resource: inwx_automated_dnssec

Automated DNSSEC management for a domain. INWX will create and manage the keys and send them to the domain registry. If you do not use INWX nameservers, use [inwx_dnssec_key](inwx_dnssec_key.md) instead.

## Example Usage

```terraform
resource "inwx_automated_dnssec" "example_com" {
  domain = "example.com"
}
```

## Argument Reference

* `domain` - (Required) Name of the domain
