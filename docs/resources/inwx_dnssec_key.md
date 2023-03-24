# Resource: inwx_dnssec_key

Provides a INWX DNSSEC key resource

## Example Usage

```terraform
resource "inwx_dnssec_key" "example_com" {
  domain = "example.com"
  public_key = "ac12c2..."
  algorithm = "SHA256"
}
```

## Argument Reference

* `domain` - (Required) Name of the domain
* `public_key` - (Required) Public key of the domain
* `algorithm` - (Required) Algorithm used for the public key

## Import

INWX DNSSEC keys can be imported using the domain name and digest e.g.,

```
$ terraform import inwx_dnssec_key.example_com example.com/4E1243BD22C66E76C2BA9EDDC1F91394E57F9F83
```

## Caveats

### Nameservers

Can not be used in conjunction with INWX nameservers.
