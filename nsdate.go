package main

import (
	"time"
	"strconv"
	"strings"
	"unicode"
)

var ns_magic_convert = map[string]string{
	"M":    "1",
	"MM":   "01",
	"MMM":  "Jan",
	"MMMM": "January",
	"d":    "2",
	"dd":   "02",
	"E":    "Mon",
	"EEEE": "Monday",
	"h":    "3",
	"hh":   "03",
	"HH":   "15",
	"a":    "PM",
	"m":    "4",
	"mm":   "04",
	"s":    "5",
	"ss":   "05",
	"SSS":  ".000",
}

// this conversion system obviously isn't
// perfect, but it supports many common
// formatters and fills in the gaps in Go's
// magic numbers to be tighter to the base
// Unicode spec

// unsupported formatters are warned to the user
// and ommitted from render.

func nsdate(t time.Time, formatter string) string {
	final := strings.Builder {}

	formatter = strings.TrimSpace(formatter)

	for {
		if len(formatter) == 0 {
			break
		}

		for _, c := range formatter {
			if unicode.IsLetter(c) {
				n      := count_rune(formatter, c)
				repeat := formatter[:n]

				// years
				if c == 'y' {
					switch n {
					case 1:
						final.WriteString(strconv.Itoa(t.Year()))
					case 2:
						final.WriteString(t.Format("06"))
					default:
						y := strconv.Itoa(t.Year())
						final.WriteString(strings.Repeat("0", clamp(n-len(y))))
						final.WriteString(y)
					}
					formatter = formatter[n:]
					break
				}

				// H - unpadded hour
				if c == 'H' && n == 1 {
					final.WriteString(strconv.Itoa(t.Hour()))
				}
				// MMMMM - single letter month
				if c == 'M' && n == 5 {
					final.WriteString(t.Month().String()[:1])
				}
				// EEEEE - single letter week
				if c == 'E' && n == 5 {
					final.WriteString(t.Weekday().String()[:1])
				}
				// EEEEEE - two letter week
				if c == 'E' && n == 6 {
					final.WriteString(t.Weekday().String()[:2])
				}

				if x, ok := ns_magic_convert[repeat]; ok {
					final.WriteString(t.Format(x))
				}

				formatter = formatter[n:]
				break
			}

			final.WriteRune(c)
			formatter = formatter[1:]
		}
	}

	return final.String()
}

func clamp(m int) int {
	if m < 0 {
		return 0
	}
	return m
}