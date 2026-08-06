// Package werkzeugpw reimplements the password-hash format used by the
// previous panel version (pbkdf2/scrypt via a salted, method-prefixed
// encoding), so it can produce and parse byte-identical output for the
// same method/salt/password against existing stored hashes, rather than
// just "a secure hash" of its own design.
//
// Ground truth for this format is the previous panel's password-hashing
// implementation. That implementation's default method has changed over
// time (from pbkdf2 to scrypt, and the pbkdf2 iteration count has also
// increased), so CheckPasswordHash still needs to parse whatever method is
// embedded in a given stored hash, since existing rows may predate either
// change.
package werkzeugpw

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // sha1 is one of the pbkdf2 hash names this format supports; needed to verify pre-existing hashes, not chosen for new ones
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"math/big"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

// saltChars is the alphabet used to generate the random per-hash salt.
const saltChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const (
	defaultSaltLength   = 16
	defaultPBKDF2Iters  = 1_000_000
	defaultScryptN      = 1 << 15
	defaultScryptR      = 8
	defaultScryptP      = 1
	scryptDerivedKeyLen = 64 // derived key length required for compatibility with existing stored hashes
)

func genSalt(length int) (string, error) {
	b := make([]byte, length)
	max := big.NewInt(int64(len(saltChars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = saltChars[n.Int64()]
	}
	return string(b), nil
}

func pbkdf2HashFunc(name string) (func() hash.Hash, int, bool) {
	switch name {
	case "sha1":
		return sha1.New, sha1.Size, true
	case "sha256":
		return sha256.New, sha256.Size, true
	case "sha384":
		return sha512.New384, sha512.Size384, true
	case "sha512":
		return sha512.New, sha512.Size, true
	default:
		return nil, 0, false
	}
}

// hashInternal derives the password hash for the given method/salt/password
// and returns it as a hex string, along with the canonical "method:args..."
// string (with defaults filled in) to embed in the stored hash.
func hashInternal(method, salt, password string) (hexDigest, canonicalMethod string, err error) {
	parts := strings.Split(method, ":")
	name, args := parts[0], parts[1:]

	switch name {
	case "scrypt":
		n, r, p := defaultScryptN, defaultScryptR, defaultScryptP
		if len(args) == 3 {
			n, err = strconv.Atoi(args[0])
			if err == nil {
				r, err = strconv.Atoi(args[1])
			}
			if err == nil {
				p, err = strconv.Atoi(args[2])
			}
			if err != nil {
				return "", "", fmt.Errorf("'scrypt' takes 3 arguments: %w", err)
			}
		} else if len(args) != 0 {
			return "", "", fmt.Errorf("'scrypt' takes 3 arguments")
		}

		key, err := scrypt.Key([]byte(password), []byte(salt), n, r, p, scryptDerivedKeyLen)
		if err != nil {
			return "", "", err
		}
		return hex.EncodeToString(key), fmt.Sprintf("scrypt:%d:%d:%d", n, r, p), nil

	case "pbkdf2":
		hashName := "sha256"
		iterations := defaultPBKDF2Iters
		switch len(args) {
		case 0:
		case 1:
			hashName = args[0]
		case 2:
			hashName = args[0]
			iterations, err = strconv.Atoi(args[1])
			if err != nil {
				return "", "", fmt.Errorf("'pbkdf2' takes 2 arguments: %w", err)
			}
		default:
			return "", "", fmt.Errorf("'pbkdf2' takes 2 arguments")
		}

		h, keyLen, ok := pbkdf2HashFunc(hashName)
		if !ok {
			return "", "", fmt.Errorf("unsupported pbkdf2 hash %q", hashName)
		}
		key := pbkdf2.Key([]byte(password), []byte(salt), iterations, keyLen, h)
		return hex.EncodeToString(key), fmt.Sprintf("pbkdf2:%s:%d", hashName, iterations), nil

	default:
		return "", "", fmt.Errorf("invalid hash method %q", name)
	}
}

// GeneratePasswordHash hashes password using the current default method
// ("scrypt") with a freshly generated random salt.
func GeneratePasswordHash(password string) (string, error) {
	salt, err := genSalt(defaultSaltLength)
	if err != nil {
		return "", err
	}
	digest, canonicalMethod, err := hashInternal("scrypt", salt, password)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s$%s$%s", canonicalMethod, salt, digest), nil
}

// CheckPasswordHash reports whether password matches pwhash. It returns
// false (not an error) for any malformed or unsupported hash, so an
// unrecognized hash format simply fails verification rather than
// panicking or bubbling up an error.
func CheckPasswordHash(pwhash, password string) bool {
	parts := strings.SplitN(pwhash, "$", 3)
	if len(parts) != 3 {
		return false
	}
	method, salt, wantHex := parts[0], parts[1], parts[2]

	gotHex, _, err := hashInternal(method, salt, password)
	if err != nil {
		return false
	}

	return hmac.Equal([]byte(strings.ToLower(gotHex)), []byte(strings.ToLower(wantHex)))
}
