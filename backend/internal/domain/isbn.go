package domain

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// IdentifierType classifies an external book identifier.
type IdentifierType string

const (
	IdentifierTypeISBN10 IdentifierType = "ISBN_10"
	IdentifierTypeISBN13 IdentifierType = "ISBN_13"
	IdentifierTypeASIN   IdentifierType = "ASIN"
)

var (
	asinRegex = regexp.MustCompile(`^[A-Z0-9]{10}$`)
)

// CleanIdentifier strips all whitespace, hyphens, and converts to uppercase.
func CleanIdentifier(raw string) string {
	var sb strings.Builder
	for _, r := range raw {
		if r != '-' && r != ' ' && !unicode.IsSpace(r) {
			sb.WriteRune(unicode.ToUpper(r))
		}
	}
	return sb.String()
}

// ValidateISBN10 validates an ISBN-10 string using Modulo 11 check digit arithmetic.
func ValidateISBN10(raw string) bool {
	cleaned := CleanIdentifier(raw)
	if len(cleaned) != 10 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		char := cleaned[i]
		if char < '0' || char > '9' {
			return false
		}
		digit := int(char - '0')
		sum += digit * (10 - i)
	}

	// Last character can be 0-9 or 'X' (representing 10)
	lastChar := cleaned[9]
	if lastChar == 'X' {
		sum += 10
	} else if lastChar >= '0' && lastChar <= '9' {
		sum += int(lastChar - '0')
	} else {
		return false
	}

	return sum%11 == 0
}

// ValidateISBN13 validates an ISBN-13 string using Modulo 10 check digit arithmetic.
func ValidateISBN13(raw string) bool {
	cleaned := CleanIdentifier(raw)
	if len(cleaned) != 13 {
		return false
	}

	// ISBN-13 must start with 978 or 979
	if !strings.HasPrefix(cleaned, "978") && !strings.HasPrefix(cleaned, "979") {
		return false
	}

	sum := 0
	for i := 0; i < 12; i++ {
		char := cleaned[i]
		if char < '0' || char > '9' {
			return false
		}
		digit := int(char - '0')
		if i%2 == 0 {
			sum += digit * 1
		} else {
			sum += digit * 3
		}
	}

	lastChar := cleaned[12]
	if lastChar < '0' || lastChar > '9' {
		return false
	}
	checkDigit := int(lastChar - '0')

	expectedCheck := (10 - (sum % 10)) % 10
	return checkDigit == expectedCheck
}

// ValidateASIN validates an Amazon Standard Identification Number (10 alphanumeric characters).
func ValidateASIN(raw string) bool {
	cleaned := CleanIdentifier(raw)
	return asinRegex.MatchString(cleaned)
}

// ISBN10To13 converts a valid ISBN-10 into its canonical ISBN-13 representation.
func ISBN10To13(isbn10 string) (string, error) {
	cleaned := CleanIdentifier(isbn10)
	if !ValidateISBN10(cleaned) {
		return "", fmt.Errorf("%w: %s", ErrInvalidISBN, isbn10)
	}

	// Prefix with 978 and take the first 9 digits of the ISBN-10
	base := "978" + cleaned[:9]

	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(base[i] - '0')
		if i%2 == 0 {
			sum += digit * 1
		} else {
			sum += digit * 3
		}
	}

	checkDigit := (10 - (sum % 10)) % 10
	return fmt.Sprintf("%s%d", base, checkDigit), nil
}

// NormalizeISBN accepts any ISBN representation (10 or 13, with or without hyphens),
// validates it, and returns the canonical ISBN-13 (and ISBN-10 if derivable).
func NormalizeISBN(raw string) (isbn13 string, isbn10 string, err error) {
	cleaned := CleanIdentifier(raw)

	if len(cleaned) == 13 {
		if !ValidateISBN13(cleaned) {
			return "", "", fmt.Errorf("%w: invalid ISBN-13 %s", ErrInvalidISBN, raw)
		}
		// If it starts with 978, we can compute the equivalent ISBN-10
		if strings.HasPrefix(cleaned, "978") {
			body := cleaned[3:12]
			sum := 0
			for i := 0; i < 9; i++ {
				sum += int(body[i]-'0') * (10 - i)
			}
			rem := (11 - (sum % 11)) % 11
			check := "X"
			if rem < 10 {
				check = fmt.Sprintf("%d", rem)
			}
			isbn10 = body + check
		}
		return cleaned, isbn10, nil
	}

	if len(cleaned) == 10 {
		if !ValidateISBN10(cleaned) {
			return "", "", fmt.Errorf("%w: invalid ISBN-10 %s", ErrInvalidISBN, raw)
		}
		isbn13, err := ISBN10To13(cleaned)
		if err != nil {
			return "", "", err
		}
		return isbn13, cleaned, nil
	}

	return "", "", fmt.Errorf("%w: expected 10 or 13 characters, got %d for %q", ErrInvalidISBN, len(cleaned), raw)
}
