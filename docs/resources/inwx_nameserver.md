# Resource: inwx_nameserver

Provides a INWX nameserver zone resource on the anycast nameserver network (50+ locations worldwide). Needed if you use INWX nameservers for [inwx_domain](inwx_domain.md). Use [inwx_nameserver_record](inwx_nameserver_record.md) to create records in the zone.

## Example Usage

```terraform
// primary
resource "inwx_nameserver" "example_com_nameserver" {
  domain = "example.com"
  type = "MASTER"
  nameservers = [
    "ns.inwx.de",
    "ns2.inwx.de"
  ]
}

// or secondary
resource "inwx_nameserver" "example_com_nameserver" {
  domain = "example.com"
  type = "SLAVE"
  master_ip = "1.2.3.4"
}
```

## Argument Reference

* `domain` - (Required) Name of the domain
* `type` - (Required) Type of the nameserver zone. One of: `MASTER`, `SLAVE`
* `nameservers` - (Required) List of nameservers
* `master_ip` - (Optional) Master IP address
* `web` - (Optional) Web nameserver entry
* `mail` - (Optional) Mail nameserver entry
* `soa_mail` - (Optional) 	Email address for SOA record
* `url_redirect_type` - (Optional) Type of the url redirection. One of: `HEADER301`, `HEADER302`, `FRAME`
* `url_redirect_title` - (Optional) Title of the frame redirection
* `url_redirect_description` - (Optional) Description of the frame redirection
* `url_redirect_fav_icon` - (Optional) FavIcon of the frame redirection
* `url_redirect_keywords` - (Optional) Keywords of the frame redirection
* `testing` - (Optional) Execute command in testing mode. Default: `false`
* `ignore_existing` - (Optional) Ignore existing. Default: `false`

## Import

INWX nameserver zones can be imported using the `id`, e.g.,

```
$ terraform import inwx_nameserver example.com:2147483647
```
