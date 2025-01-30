terraform {
  required_providers {
    inwx = {
      source  = "inwx/inwx"
      version = ">= 1.0.0"
    }
  }
}

// API configuration
provider "inwx" {
  api_url  = "https://api.ote.domrobot.com/jsonrpc/"
  username = "example-user"
  password = "redacted"
  tan      = "000000"
}

// contact used for domains
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

// manage domains
resource "inwx_domain" "example_com" {
  name = "example.com"
  nameservers = [
    // if you want to use inwx ns, create a zone with inwx_nameserver
    "ns.inwx.de",
    "ns2.inwx.de"
  ]
  period        = "1Y"
  renewal_mode  = "AUTORENEW"
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

// zone in anycast dns
resource "inwx_nameserver" "example_com_nameserver" {
  domain = "example.com"
  type   = "MASTER"
  nameservers = [
    "ns.inwx.de",
    "ns2.inwx.de"
  ]
}

// nameserver record for a zone from above
resource "inwx_nameserver_record" "example_com_txt_1" {
  domain  = "example.com"
  type    = "TXT"
  content = "DNS records with terraform"
}

// dnssec when inwx nameservers are used
resource "inwx_automated_dnssec" "example_com" {
  domain = "example.com"
}

// dnssec for external nameservers
resource "inwx_dnssec_key" "example_com" {
  domain     = "example.com"
  public_key = "ac12c2..."
  algorithm  = "SHA256"
}

// glue record
resource "inwx_glue_record" "example_com_glue_1" {
  hostname = "example.com"
  ip = [
    "192.168.0.1"
  ]
}