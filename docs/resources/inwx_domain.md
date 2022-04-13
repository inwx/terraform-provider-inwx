# Resource: inwx_domain

Description of what this resource does, with links to official
app/service documentation.

## Example Usage

```terraform
resource "inwx_domain" "example_com" {
  name = "example.com"
  nameservers = [
    "ns.inwx.de",
    "ns2.inwx.de"
  ]
  period = "1Y"
  renewal_mode = "AUTOEXPIRE"
  transfer_lock = true
  contacts {
    registrant = 2147483647 // id of contact
    admin  = 2147483647 // id of contact
    tech  = 2147483647 // id of contact
    billing  = 2147483647 // id of contact
  }
  extra_data = {
    // Enable e.g. whois protection
    "WHOIS-CURRENCY": "EUR",
    "WHOIS-PROTECTION": "1" // 1 == bool true
  }
}
```

### Full Example With Terraform Managed Contacts

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

resource "inwx_domain" "example_com" {
  name = "example.com"
  nameservers = [
    "ns.inwx.de",
    "ns2.inwx.de"
  ]
  period = "1Y"
  renewal_mode = "AUTOEXPIRE"
  transfer_lock = true
  contacts {
    // references to terraform managed contact "example_person"
    registrant = inwx_domain_contact.example_person.id
    admin  = inwx_domain_contact.example_person.id
    tech  = inwx_domain_contact.example_person.id
    billing  = inwx_domain_contact.example_person.id
  }
  extra_data = {
    // Enable e.g. whois protection
    "WHOIS-CURRENCY": "EUR",
    "WHOIS-PROTECTION": "1" // 1 == bool true
  }
}
```

## Argument Reference

* `name` - (Required) Name of the domain
* `nameservers` - (Required) Set of nameservers of the domain. Min Items: 1
* `period` - (Required) Registration period of the domain. Valid types: https://www.inwx.de/en/help/apidoc/f/ch03.html#type.period
* `renewal_mode` - (Optional) Renewal mode of the domain. One of: `AUTORENEW`, `AUTODELETE`, `AUTOEXPIRE`. Default: `AUTORENEW`
* `transfer_lock` - (Optional) Whether the domain transfer lock should be enabled. Default: `true`
* `contacts` - (Required) Contacts of the domain
* `extra_data` - (Optional) Extra data, needed for some jurisdictions. Valid extra data types: https://www.inwx.de/en/help/apidoc/f/ch03.html#type.extdata

### Nested Fields

`contacts`
* `registrant` - (Required) Id of the registrant contact
* `admin` - (Required) Id of the admin contact
* `tech` - (Required) Id of the tech contact
* `billing` - (Required) Id of the billing contact

## Attribute Reference

* `id` - Name of the domain

## Import

INWX Domains can be imported using the domain name, e.g.,

```
$ terraform import inwx_domain.example_com "example.com"
```