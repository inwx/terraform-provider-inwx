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
