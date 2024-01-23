# Resource: inwx_nameserver

Provides a INWX nameserver zone resource

## Example Usage

```terraform
resource "inwx_nameserver" "example_com_nameserver" {
  domain = "example.com"
  type = "MASTER"
  nameservers = [
    "ns.inwx.de",
    "ns2.inwx.de"
  ]
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

INWX nameserver records can be imported using the `id`, e.g.,

```
$ terraform import inwx_nameserver example.com:2147483647
```
