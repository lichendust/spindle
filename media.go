package main

import (
	"fmt"
	"strings"
	"unicode"
	"strconv"
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

	viewcode := args[0]
	args = args[1:]

	// @todo ratio currently busted, need to come up with better recommendation for userland CSS
	ratio := ""

	if len(args) > 1 {
		val := split_rune(args[0], ':')

		if len(val) > 1 {
			x, err := strconv.ParseFloat(val[0], 32); if err != nil { panic(err) }
			y, err := strconv.ParseFloat(val[1], 32); if err != nil { panic(err) }

			ratio = fmt.Sprintf(` style="padding-top: %.2f%%"`, y / x * 100.0)
		}

		args = args[1:]
	}

	switch service {
		case MEDIA_YOUTUBE: return media_youtube(viewcode, ratio, args)
		case MEDIA_VIMEO:   return media_vimeo(viewcode,   ratio, args)
	}

	return ""
}

const (
	MEDIA_YOUTUBE uint8 = iota
	MEDIA_VIMEO
)

const (
	media_vimeo_template   = `<div class='video'%s><iframe src='https://player.vimeo.com/video/%s?color=0&title=0&byline=0&portrait=0' frameborder='0' allow='fullscreen' allowfullscreen></iframe></div>`
	media_youtube_template = `<div class='video'%s><iframe src='https://www.youtube-nocookie.com/embed/%s?rel=0&controls=1' frameborder='0' allow='accelerometer; encrypted-media; gyroscope; picture-in-picture' allowfullscreen></iframe></div>`
)

func media_vimeo(viewcode, ratio string, args []string) string {
	iframe := sprint(media_vimeo_template, ratio, viewcode)

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

func media_youtube(viewcode, ratio string, args []string) string {
	iframe := sprint(media_youtube_template, ratio, viewcode)

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