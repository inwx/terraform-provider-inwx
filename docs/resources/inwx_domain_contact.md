# Resource: inwx_domain_contact

Provides a INWX domain contact resource. Needed for [inwx_domain](inwx_domain.md).

## Example Usage

```terraform
resource "inwx_domain_contact" "example_person" {
  type = "PERSON"
  name = "Example Person"
  street_address = "Example Street 0"
  city = "Example City"
  postal_code = 00000
  state_province = "Example State"
  country_code = "EX"
  phone_number = "+00.00000000000"
  email = "person@example.invalid"
}
```

## Argument Reference

* `type` - (Required) Type of contact. One of: `ORG`, `PERSON`, `ROLE`
* `name` - (Required) First and lastname of the contact
* `organization` - (Optional) The legal name of the organization. Required for types other than person
* `street_address` - (Required) Street Address of the contact
* `city` - (Required) City of the contact
* `postal_code` - (Required) Postal Code/Zipcode of the contact
* `state_province` - (Optional) State/Province name of the contact. Required for certain TLDs
* `country_code` - (Required) Country code of the contact. Must be two characters
* `phone_number` - (Required) Phone number of the contact
* `fax` - (Optional) Fax number of the contact
* `email` - (Required) Contact email address
* `remarks` - (Optional) Custom description of the contact

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Id of the contact

## Import

INWX domain contacts can be imported using the `id`, e.g.,

```
$ terraform import inwx_domain_contact 2147483647
```
