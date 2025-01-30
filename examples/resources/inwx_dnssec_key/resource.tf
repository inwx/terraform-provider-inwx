resource "inwx_dnssec_key" "example_com" {
  domain     = "example.com"
  public_key = "ac12c2..."
  algorithm  = "SHA256"
}
