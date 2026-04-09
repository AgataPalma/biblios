// internal/isbn/isbn.go
package isbn

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Clean removes dashes, spaces, and other formatting from ISBN
func Clean(isbn string) string {
	re := regexp.MustCompile(`[^0-9Xx]`)
	return strings.ToUpper(re.ReplaceAllString(isbn, ""))
}

// IsValid checks if an ISBN is valid (either ISBN-10 or ISBN-13)
func IsValid(isbn string) bool {
	cleaned := Clean(isbn)
	return IsValidISBN10(cleaned) || IsValidISBN13(cleaned)
}

// IsValidISBN10 validates ISBN-10 format and check digit
func IsValidISBN10(isbn string) bool {
	if len(isbn) != 10 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		if isbn[i] < '0' || isbn[i] > '9' {
			return false
		}
		digit := int(isbn[i] - '0')
		sum += (10 - i) * digit
	}

	// Check digit can be 0-9 or X (representing 10)
	lastChar := isbn[9]
	var checkDigit int
	if lastChar == 'X' {
		checkDigit = 10
	} else if lastChar >= '0' && lastChar <= '9' {
		checkDigit = int(lastChar - '0')
	} else {
		return false
	}

	return (sum+checkDigit)%11 == 0
}

// IsValidISBN13 validates ISBN-13 format and check digit
func IsValidISBN13(isbn string) bool {
	if len(isbn) != 13 {
		return false
	}

	sum := 0
	for i := 0; i < 12; i++ {
		if isbn[i] < '0' || isbn[i] > '9' {
			return false
		}
		digit := int(isbn[i] - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	if isbn[12] < '0' || isbn[12] > '9' {
		return false
	}
	checkDigit := int(isbn[12] - '0')

	return (10-(sum%10))%10 == checkDigit
}

// ConvertToISBN13 converts ISBN-10 to ISBN-13
func ConvertToISBN13(isbn10 string) (string, error) {
	cleaned := Clean(isbn10)
	if !IsValidISBN10(cleaned) {
		return "", fmt.Errorf("invalid ISBN-10: %s", isbn10)
	}

	// Remove check digit from ISBN-10
	base := cleaned[:9]

	// Add 978 prefix
	isbn13Base := "978" + base

	// Calculate ISBN-13 check digit
	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(isbn13Base[i] - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	checkDigit := (10 - (sum % 10)) % 10
	return isbn13Base + strconv.Itoa(checkDigit), nil
}

// ConvertToISBN10 converts ISBN-13 to ISBN-10 (only works for 978-prefixed ISBN-13)
func ConvertToISBN10(isbn13 string) (string, error) {
	cleaned := Clean(isbn13)
	if !IsValidISBN13(cleaned) {
		return "", fmt.Errorf("invalid ISBN-13: %s", isbn13)
	}

	// Only 978-prefixed ISBN-13 can be converted to ISBN-10
	if !strings.HasPrefix(cleaned, "978") {
		return "", fmt.Errorf("cannot convert ISBN-13 with prefix other than 978")
	}

	// Extract the middle 9 digits
	base := cleaned[3:12]

	// Calculate ISBN-10 check digit
	sum := 0
	for i := 0; i < 9; i++ {
		digit := int(base[i] - '0')
		sum += (10 - i) * digit
	}

	checkDigit := (11 - (sum % 11)) % 11
	var checkChar string
	if checkDigit == 10 {
		checkChar = "X"
	} else {
		checkChar = strconv.Itoa(checkDigit)
	}

	return base + checkChar, nil
}

// ISBNPair holds both ISBN-10 and ISBN-13 when possible
type ISBNPair struct {
	ISBN10 *string `json:"isbn10,omitempty"`
	ISBN13 string  `json:"isbn13"`
}

// Normalize ensures we have both ISBN-10 and ISBN-13 when possible
func Normalize(input string) (ISBNPair, error) {
	cleaned := Clean(input)

	if IsValidISBN13(cleaned) {
		pair := ISBNPair{ISBN13: cleaned}
		if isbn10, err := ConvertToISBN10(cleaned); err == nil {
			pair.ISBN10 = &isbn10
		}
		return pair, nil
	}

	if IsValidISBN10(cleaned) {
		isbn13, err := ConvertToISBN13(cleaned)
		if err != nil {
			return ISBNPair{}, err
		}
		return ISBNPair{
			ISBN10: &cleaned,
			ISBN13: isbn13,
		}, nil
	}

	return ISBNPair{}, fmt.Errorf("invalid ISBN format: %s", input)
}
