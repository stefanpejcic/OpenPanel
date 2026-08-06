package cpanelpw

import "testing"

// TestVerifySHA512CryptAgainstRealPythonHash cross-checks against a hash
// actually produced by the system's SHA-512 crypt(3) implementation,
// generated with:
//
//	python3 -c "import crypt; print(crypt.crypt('mysecret123', crypt.mksalt(crypt.METHOD_SHA512)))"
//
// -> $6$viStI4RiT06JI0qH$0NWpLWJSPD71mXoQkpAvDNpFL8QXXQzqvqQr9VnKw7CB.9LUk2nEdZhz.eDZSVd/z/BBturvjsW26uTYYrH2n/
func TestVerifySHA512CryptAgainstRealPythonHash(t *testing.T) {
	hash := "$6$viStI4RiT06JI0qH$0NWpLWJSPD71mXoQkpAvDNpFL8QXXQzqvqQr9VnKw7CB.9LUk2nEdZhz.eDZSVd/z/BBturvjsW26uTYYrH2n/"

	if !VerifySHA512Crypt("mysecret123", hash) {
		t.Error("expected the real Python crypt.crypt SHA-512 hash to verify against its correct password")
	}
	if VerifySHA512Crypt("wrong-password", hash) {
		t.Error("expected the hash to reject an incorrect password")
	}
}
