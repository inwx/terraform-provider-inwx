resource "inwx_domain_contact" "example_person" {
  type           = "PERSON"
  name           = "Example Person"
  street_address = "Example Street 0"
  city           = "Example City"
  postal_code    = 00000
  state_province = "Example State"
  country_code   = "EX"
  phone_number   = "+00.00000000000"
  email          = "person@example.invalid"
}

resource "inwx_domain" "example_com" {
  name = "example.com"
  nameservers = [
    // if you want to use inwx ns, create a zone with inwx_nameserver
    "ns.inwx.de",
    "ns2.inwx.de"
  ]
  period        = "1Y"
  renewal_mode  = "AUTOEXPIRE"
  transfer_lock = true
  contacts {
    // references to terraform managed contact "example_person"
    registrant = inwx_domain_contact.example_person.id
    admin      = inwx_domain_contact.example_person.id
    tech       = inwx_domain_contact.example_person.id
    billing    = inwx_domain_contact.example_person.id
  }
  extra_data = {
    // Enable whois proxy, trustee or provide data like company number if needed
    //"WHOIS-PROTECTION": "1",
    //"ACCEPT-TRUSTEE-TAC": "1",
    //"COMPANY-NUMBER": "123",
  }
}