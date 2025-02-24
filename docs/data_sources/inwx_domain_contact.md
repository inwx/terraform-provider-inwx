# Data Source: inwx_domain_contact

Provides a INWX domain contact resource. Needed for [inwx_domain](inwx_domain.md).

## Example Usage

```terraform
data "inwx_domain_contact" "example_person" {
  id = 1
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

output "inwx_domain_contact" {
  value = data.inwx_domain_contact.example_person
}
```

## Argument Reference

* `id` - (Required) Numerical id of domain contact resource
* `type` - (Optional) Type of contact. One of: `ORG`, `PERSON`, `ROLE`
* `name` - (Optional) First and lastname of the contact
* `organization` - (Optional) The legal name of the organization. Required for types other than person
* `street_address` - (Optional) Street Address of the contact
* `city` - (Optional) City of the contact
* `postal_code` - (Optional) Postal Code/Zipcode of the contact
* `state_province` - (Optional) State/Province name of the contact
* `country_code` - (Optional) Country code of the contact. Must be two characters
* `phone_number` - (Optional) Phone number of the contact
* `fax` - (Optional) Fax number of the contact
* `email` - (Optional) Contact email address
* `remarks` - (Optional) Custom description of the contact
