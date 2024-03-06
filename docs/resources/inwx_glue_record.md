# Resource: inwx_glue_record

Provides a INWX glue record resource.

## Example Usage

```terraform
resource "inwx_glue_record" "example_com_glue_1" {
  hostname = "example.com"
  ip = [
    "192.168.0.1"
  ]
}
```

## Argument Reference

* `hostname` - (Required) Name of host
* `ip` - (Required) Ip address(es)
* `testing` - (Optional) Execute command in testing mode. Default: `false`

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Id of the glue record

## Import

INWX glue records can be imported using the `id`, e.g.,

```
$ terraform import inwx_glue_record example.com:2147483647
```
