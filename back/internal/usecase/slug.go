package usecase

import (
	"regexp"
	"strings"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

var translitMap = map[rune]string{
	'ا': "a", 'آ': "a", 'أ': "a", 'إ': "e",
	'ب': "b",
	'پ': "p",
	'ت': "t",
	'ث': "s",
	'ج': "j",
	'چ': "ch",
	'ح': "h",
	'خ': "kh",
	'د': "d",
	'ذ': "z",
	'ر': "r",
	'ز': "z",
	'ژ': "zh",
	'س': "s",
	'ش': "sh",
	'ص': "s",
	'ض': "z",
	'ط': "t",
	'ظ': "z",
	'ع': "a",
	'غ': "gh",
	'ف': "f",
	'ق': "gh",
	'ك': "k", 'ک': "k",
	'گ': "g",
	'ل': "l",
	'م': "m",
	'ن': "n",
	'و': "v",
	'ؤ': "o",
	'ه': "h", 'ة': "h", 'ۀ': "e",
	'ی': "y", 'ي': "y", 'ئ': "y",
	'ء': "",
	'۰': "0", '۱': "1", '۲': "2", '۳': "3", '۴': "4", '۵': "5", '۶': "6", '۷': "7", '۸': "8", '۹': "9",
}

func slugify(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	for _, r := range value {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			builder.WriteRune(r)
		case r == ' ' || r == '_' || r == '-' || r == 'ـ':
			builder.WriteRune('-')
		case translitMap[r] != "":
			builder.WriteString(translitMap[r])
		default:
			builder.WriteRune('-')
		}
	}

	value = slugRegex.ReplaceAllString(builder.String(), "-")
	value = strings.Trim(value, "-")
	return value
}
