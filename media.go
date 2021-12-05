package main

import (
	"strings"
	"unicode"
)

const (
	MEDIA_YOUTUBE uint8 = iota
	MEDIA_VIMEO
)

func do_media(args []string) string {
	service := uint8(255)

	switch args[0] {
	case "youtube":
		service = MEDIA_YOUTUBE
		args = args[1:]
	case "vimeo":
		service = MEDIA_VIMEO
		args = args[1:]
	}

	if !(service < 255) {
		letter := false

		for _, r := range args[0] {
			if unicode.IsLetter(r) {
				letter = true
				break
			}
		}

		if letter {
			service = MEDIA_YOUTUBE
		} else {
			service = MEDIA_VIMEO
		}
	}

	switch service {
		case MEDIA_YOUTUBE: return media_youtube(args[0], args[1:])
		case MEDIA_VIMEO:   return media_vimeo  (args[0], args[1:])
	}

	return ""
}

func media_vimeo(viewcode string, args []string) string {
	iframe := sprint(media_vimeo_template, viewcode)

	if len(args) == 0 {
		return iframe
	}

	for _, a := range args {
		if a[0] == '#' {
			iframe = strings.Replace(iframe, `color=0`, `color=` + a[1:len(a)], 1)
			continue
		}

		switch a {
		case "hide_all":       iframe = strings.Replace(iframe, `&title=0&byline=0&portrait=0`, ``, 1)
		case "hide_title":     iframe = strings.Replace(iframe, `&title=0`,    ``, 1)
		case "hide_portrait":  iframe = strings.Replace(iframe, `&portrait=0`, ``, 1)
		case "hide_byline":    iframe = strings.Replace(iframe, `&byline=0`,   ``, 1)
		}
	}

	return iframe
}

func media_youtube(viewcode string, args []string) string {
	iframe := sprint(media_youtube_template, viewcode)

	if len(args) == 0 {
		return iframe
	}

	for _, a := range args {
		switch a {
		case "hide_controls": iframe = strings.Replace(iframe, `&controls=1`, `&controls=0`, 1)
		}
	}

	return iframe
}