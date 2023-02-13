# Resource: inwx_nameserver_record

Provides a INWX nameserver record resource

## Example Usage

```terraform
resource "inwx_nameserver_record" "example_com_txt_1" {
  domain = "example.com"
  type = "TXT"
  content = "DNS records with terraform"
}
```

## Argument Reference

* `domain` - (Required) Name of the domain
* `type` - (Required) Type of the nameserver record. One of: `A`, `AAAA`, `AFSDB`, `ALIAS`, `CAA`, `CERT`, `CNAME`, 
`HINFO`, `KEY`, `LOC`, `MX`, `NAPTR`, `NS`, `OPENPGPKEY`, `PTR`, `RP`, `SMIMEA`, `SOA`, `SRV`, `SSHFP`, `TLSA`, `TXT`, 
`URI`, `URL`
* `ro_id` - (Optional) DNS domain id
* `content` - (Required) Content of the nameserver record
* `name` - (Optional) Name of the nameserver record
* `ttl` - (Optional) TTL (time to live) of the nameserver record. Default: `3600`
* `prio` - (Optional) Priority of the nameserver record. Default: `0`
* `url_redirect_type` - (Optional) Type of the url redirection. One of: `HEADER301`, `HEADER302`, `FRAME`
* `url_redirect_title` - (Optional) Title of the frame redirection
* `url_redirect_description` - (Optional) Description of the frame redirection
* `url_redirect_fav_icon` - (Optional) FavIcon of the frame redirection
* `url_redirect_keywords` - (Optional) Keywords of the frame redirection
* `url_append` - (Optional) Append the path for redirection. Default: `false`
* `testing` - (Optional) Execute command in testing mode. Default: `false`

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Id of the nameserver record

## Import

INWX nameserver records can be imported using the `id`, e.g.,

```
$ terraform import inwx_nameserver_record example.com:2147483647
```
