package rsagen

import "testing"

func TestRSAGen(t *testing.T) {
	// Create the keys
	priv, pub := GenerateRsaKeyPair(2048)
	// Export the keys to pem string
	priv_pem := ExportRsaPrivateKeyAsPem(priv)
	pub_pem, _ := ExportRsaPublicKeyAsPem(pub)

	// Import the keys from pem string
	priv_parsed, _ := ParseRsaPrivateKeyFromPem(priv_pem)
	pub_parsed, _ := ParseRsaPublicKeyFromPem(pub_pem)

	// Export the newly imported keys
	priv_parsed_pem := ExportRsaPrivateKeyAsPem(priv_parsed)
	pub_parsed_pem, _ := ExportRsaPublicKeyAsPem(pub_parsed)

	// Check that the exported/imported keys match the original keys
	if string(priv_pem) != string(priv_parsed_pem) || string(pub_pem) != string(pub_parsed_pem) {
		t.Fail()
	}
}
