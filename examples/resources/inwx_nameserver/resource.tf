// primary
resource "inwx_nameserver" "example_com_nameserver" {
  domain = "example.com"
  type   = "MASTER"
  nameservers = [
    "ns.inwx.de",
    "ns2.inwx.de"
  ]
}

// or secondary
resource "inwx_nameserver" "example_com_nameserver" {
  domain    = "example.com"
  type      = "SLAVE"
  master_ip = "1.2.3.4"
}
