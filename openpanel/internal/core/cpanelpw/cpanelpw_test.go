package cpanelpw

import "testing"

// cross-checks against a hash from the real crypt(3), reproducible with:
//
//	python3 -c "import crypt; print(crypt.crypt('mysecret123', crypt.mksalt(crypt.METHOD_SHA512)))"
func TestVerifySHA512CryptAgainstRealHash(t *testing.T) {
	hash := "$6$viStI4RiT06JI0qH$0NWpLWJSPD71mXoQkpAvDNpFL8QXXQzqvqQr9VnKw7CB.9LUk2nEdZhz.eDZSVd/z/BBturvjsW26uTYYrH2n/"

	if !VerifySHA512Crypt("mysecret123", hash) {
		t.Error("expected the real crypt(3) SHA-512 hash to verify against its correct password")
	}
	if VerifySHA512Crypt("wrong-password", hash) {
		t.Error("expected the hash to reject an incorrect password")
	}
}
