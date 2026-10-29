package projectverify

import (
	"encoding/base64"
	"strings"
)

const (
	hexGUIDLength                   = 32
	tuanjieEncodedGUIDLength        = 56
	tuanjieEncodedGUIDDecodedLength = 41
)

// tuanjieGUIDPrefixKey turns the first decoded bytes of a Tuanjie-encoded GUID into lower-case hex
// digits. Tuanjie publishes no spec for the encoding; the key was found empirically and holds for
// every encoded GUID of a Tuanjie 1.9.1 project (4578 of them) and of the com.tuanjie.particle-system
// package bundled with Tuanjie 2022.3.48t7. It only checks the value: the rest of the encoding is
// unknown, so an encoded GUID cannot be turned back into its 32-hex form.
var tuanjieGUIDPrefixKey = [...]byte{0x3f, 0x4a, 0x2f, 0x85, 0x1a, 0x9d, 0x63, 0x4b}

// guidForm is how a guid value is written.
type guidForm int

const (
	guidFormInvalid guidForm = iota
	guidFormHex
	guidFormTuanjieEncoded
	// guidFormTuanjieEncodedMismatch has the shape of a Tuanjie-encoded GUID but fails the prefix check.
	guidFormTuanjieEncodedMismatch
)

// classifyGUID returns the form of value and, for a valid value, the key duplicates are compared
// by. Hex keys are lower-cased because Unity reads hex in either case; encoded keys are kept as
// written because Base64 is case-sensitive. The keys of the two forms differ in length, so a hex
// GUID and an encoded GUID never share a key.
func classifyGUID(value string) (string, guidForm) {
	if isValidHexGUID(value) {
		return strings.ToLower(value), guidFormHex
	}
	decoded, ok := decodeTuanjieEncodedShape(value)
	if !ok {
		return "", guidFormInvalid
	}
	if !hasTuanjieGUIDPrefix(decoded) {
		return "", guidFormTuanjieEncodedMismatch
	}
	return value, guidFormTuanjieEncoded
}

// isValidHexGUID reports whether value is 32 hexadecimal characters that are not all zero.
func isValidHexGUID(value string) bool {
	if len(value) != hexGUIDLength {
		return false
	}
	allZero := true
	for _, character := range value {
		if !isHexDigit(character) {
			return false
		}
		if character != '0' {
			allZero = false
		}
	}
	return !allZero
}

func isHexDigit(character rune) bool {
	return ('0' <= character && character <= '9') ||
		('a' <= character && character <= 'f') ||
		('A' <= character && character <= 'F')
}

// decodeTuanjieEncodedShape decodes value when it is 55 standard Base64 characters followed by
// one "=" and decodes to 41 bytes.
func decodeTuanjieEncodedShape(value string) ([]byte, bool) {
	if len(value) != tuanjieEncodedGUIDLength || value[len(value)-1] != '=' {
		return nil, false
	}
	for i := 0; i < len(value)-1; i++ {
		if !isBase64Character(value[i]) {
			return nil, false
		}
	}
	// Not strict: Tuanjie's padding bits are not known to be zero.
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil || len(decoded) != tuanjieEncodedGUIDDecodedLength {
		return nil, false
	}
	return decoded, true
}

func isLowerHexDigit(character byte) bool {
	return ('0' <= character && character <= '9') || ('a' <= character && character <= 'f')
}

func isBase64Character(character byte) bool {
	return ('A' <= character && character <= 'Z') ||
		('a' <= character && character <= 'z') ||
		('0' <= character && character <= '9') ||
		character == '+' || character == '/'
}

func hasTuanjieGUIDPrefix(decoded []byte) bool {
	for i, key := range tuanjieGUIDPrefixKey {
		if !isLowerHexDigit(decoded[i] ^ key) {
			return false
		}
	}
	return true
}
