// Package cpanelpw verifies $6$ (SHA-512 crypt) password hashes, the
// format cPanel-imported user accounts carry until their first successful
// login upgrades them to the panel's own password hash format.
package cpanelpw

import "github.com/GehirnInc/crypt/sha512_crypt"

// VerifySHA512Crypt reports whether password matches storedHash by
// recomputing the SHA-512 crypt digest with the salt embedded in
// storedHash and comparing the result.
func VerifySHA512Crypt(password, storedHash string) bool {
	return sha512_crypt.New().Verify(storedHash, []byte(password)) == nil
}
