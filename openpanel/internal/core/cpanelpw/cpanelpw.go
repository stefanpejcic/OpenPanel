// Package cpanelpw verifies $6$ (SHA-512 crypt) hashes, the format cPanel-imported accounts carry until their first login upgrades them to our own hash format.
package cpanelpw

import "github.com/GehirnInc/crypt/sha512_crypt"

// VerifySHA512Crypt reports whether password matches storedHash, using the salt embedded in storedHash
func VerifySHA512Crypt(password, storedHash string) bool {
	return sha512_crypt.New().Verify(storedHash, []byte(password)) == nil
}
