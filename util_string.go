package qore

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// StringToCammelCase converts given string to the cammel case.
func StringToCammelCase(s string) string {
	var result []rune
	shouldCapitalize := false
	for i, char := range s {
		if i == 0 {
			shouldCapitalize = true
		}
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && !unicode.IsSpace(char) {
			shouldCapitalize = true
		} else {
			if shouldCapitalize {
				result = append(result, unicode.ToUpper(char))
				shouldCapitalize = false
			} else {
				result = append(result, char)
			}
		}
	}

	return string(result)
}

// StringToSnakeCase converts given string to the snake case.
func StringToSnakeCase(s string) string {
	var result []rune
	for i, char := range s {
		if unicode.IsUpper(char) {
			if i > 0 && result[len(result)-1] != '_' {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(char))
		} else if unicode.IsLetter(char) || unicode.IsDigit(char) {
			result = append(result, char)
		} else if unicode.IsSpace(char) || (!unicode.IsLetter(char) && !unicode.IsDigit(char)) {
			if len(result) > 0 && result[len(result)-1] != '_' {
				result = append(result, '_')
			}
		}
	}

	finalResult := strings.TrimRight(string(result), "_")
	return finalResult
}

// StringRemoveNonAlphabet removes non-alphabet characters from given string.
func StringRemoveNonAlphabet(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// StringIsFirstCharNonAlphabet checks is first character from given string contains non-alphabet.
func StringIsFirstCharNonAlphabet(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'))
}

const (
	randomAlphaNumCharset    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	randomAlphaNumCharsetLen = len(randomAlphaNumCharset)
	randomAlphaNumMaxByte    = 255 - (256 % randomAlphaNumCharsetLen)
)

var randomReaderPool = sync.Pool{New: func() any {
	return bufio.NewReader(rand.Reader)
}}

// StringAlphaNumRandom generate alpha numeric random string.
// It will return random string and an error if failed to read the bytes.
func StringAlphaNumRandom(length uint8) (string, error) {
	reader := randomReaderPool.Get().(*bufio.Reader)
	defer randomReaderPool.Put(reader)

	b := make([]byte, length)
	r := make([]byte, length+(length/4))
	var i uint8 = 0

	for {
		_, err := io.ReadFull(reader, r)
		if err != nil {
			return "", fmt.Errorf("failed to read buffer: %w", err)
		}
		for _, rb := range r {
			if rb > byte(randomAlphaNumMaxByte) {
				continue
			}
			b[i] = randomAlphaNumCharset[rb%byte(randomAlphaNumCharsetLen)]
			i++
			if i == length {
				return string(b), nil
			}
		}
	}
}
