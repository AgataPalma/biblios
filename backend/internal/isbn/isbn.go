package isbn

import (
	"fmt"
	"strings"
)

type Pair struct {
	ISBN10 *string
	ISBN13 string
}

// Clean strips hyphens and spaces from an ISBN string.
func Clean(s string) string {
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	return strings.TrimSpace(s)
}

// Normalize validates and normalises an ISBN-10 or ISBN-13 string.
// It returns a Pair with both forms where possible.
func Normalize(raw string) (Pair, error) {
	s := Clean(raw)
	switch len(s) {
	case 10:
		if !validISBN10(s) {
			return Pair{}, fmt.Errorf("invalid ISBN-10: %s", raw)
		}
		isbn13 := isbn10to13(s)
		return Pair{ISBN10: &s, ISBN13: isbn13}, nil
	case 13:
		if !validISBN13(s) {
			return Pair{}, fmt.Errorf("invalid ISBN-13: %s", raw)
		}
		isbn10 := isbn13to10(s)
		return Pair{ISBN10: isbn10, ISBN13: s}, nil
	default:
		return Pair{}, fmt.Errorf("invalid ISBN length (%d): %s", len(s), raw)
	}
}

func validISBN10(s string) bool {
	sum := 0
	for i, c := range s[:9] {
		if c < '0' || c > '9' {
			return false
		}
		sum += int(c-'0') * (10 - i)
	}
	last := s[9]
	if last == 'X' || last == 'x' {
		sum += 10
	} else if last >= '0' && last <= '9' {
		sum += int(last - '0')
	} else {
		return false
	}
	return sum%11 == 0
}

func validISBN13(s string) bool {
	sum := 0
	for i, c := range s {
		if c < '0' || c > '9' {
			return false
		}
		d := int(c - '0')
		if i%2 == 0 {
			sum += d
		} else {
			sum += d * 3
		}
	}
	return sum%10 == 0
}

func isbn10to13(s string) string {
	base := "978" + s[:9]
	sum := 0
	for i, c := range base {
		d := int(c - '0')
		if i%2 == 0 {
			sum += d
		} else {
			sum += d * 3
		}
	}
	check := (10 - sum%10) % 10
	return fmt.Sprintf("%s%d", base, check)
}

func isbn13to10(s string) *string {
	if !strings.HasPrefix(s, "978") {
		return nil
	}
	base := s[3:12]
	sum := 0
	for i, c := range base {
		sum += int(c-'0') * (10 - i)
	}
	check := (11 - sum%11) % 11
	var last string
	if check == 10 {
		last = "X"
	} else {
		last = fmt.Sprintf("%d", check)
	}
	result := base + last
	return &result
}
