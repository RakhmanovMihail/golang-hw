package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(str string) (string, error) {
	var builder strings.Builder
	var previousRune rune
	markNextAsRune := false
	for _, r := range str {
		switch {
		case previousRune == rune(0) && unicode.IsDigit(r):
			return "", ErrInvalidString
		case unicode.IsDigit(r) && !markNextAsRune:
			count, err := strconv.Atoi(string(r))
			if err != nil {
				return "", err
			}
			builder.WriteString(strings.Repeat(string(previousRune), count))
			previousRune = rune(0)
		case r == '\\' && !markNextAsRune:
			markNextAsRune = true
		default:
			if unicode.IsLetter(r) && markNextAsRune {
				return "", ErrInvalidString
			}
			if previousRune != rune(0) {
				builder.WriteRune(previousRune)
			}
			markNextAsRune = false
			previousRune = r
		}
	}
	if markNextAsRune {
		return "", ErrInvalidString
	}
	if previousRune != rune(0) {
		builder.WriteRune(previousRune)
	}
	return builder.String(), nil
}
