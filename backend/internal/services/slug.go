package services

import (
	"regexp"
	"strings"
	"unicode"
)

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func Slug(s string) string {
	r := strings.NewReplacer("č", "c", "ć", "c", "š", "s", "ž", "z", "đ", "dj", "Č", "c", "Ć", "c", "Š", "s", "Ž", "z", "Đ", "dj").Replace(s)
	r = strings.Map(func(x rune) rune {
		if unicode.IsLetter(x) || unicode.IsDigit(x) {
			return unicode.ToLower(x)
		}
		return '-'
	}, r)
	return strings.Trim(nonSlug.ReplaceAllString(r, "-"), "-")
}
