package frame

import (
	"strings"
)

// Encodes a header value using STOMP value encoding
// TODO: replace with more efficient version.
func encodeValue(s string) []byte {
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "\r", "\\r", -1)
	s = strings.Replace(s, "\n", "\\n", -1)
	s = strings.Replace(s, ":", "\\c", -1)
	return []byte(s)
}

// Unencodes a header value using STOMP value encoding
// TODO: replace with more efficient version.
// TODO: return error if invalid sequences found (eg "\t")
func unencodeValue(b []byte) (string, error) {
	s := string(b)
	s = strings.Replace(s, "\\r", "\r", -1)
	s = strings.Replace(s, "\\n", "\n", -1)
	s = strings.Replace(s, "\\c", ":", -1)
	s = strings.Replace(s, "\\\\", "\\", -1)
	return s, nil
}
