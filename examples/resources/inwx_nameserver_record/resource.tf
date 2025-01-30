resource "inwx_nameserver_record" "example_com_txt_1" {
  domain  = "example.com"
  type    = "TXT"
  content = "DNS records with terraform"
}
