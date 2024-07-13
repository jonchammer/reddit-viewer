package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"
)

// NOTE: Our assets are embedded in the binary to ensure that they are always
// available, regardless of which directory the application is running in. This
// also helps to simplify testing, since the embedded files are accessible at
// both test time and run time.
//
//go:embed templates
var templates embed.FS

func renderFeed(feed *Feed) ([]byte, error) {

	tmpl, err := template.New("feed.html").Funcs(template.FuncMap{
		"formatTime": formatTimeSincePost,
		"typeString": typeString,
	}).ParseFS(templates, "templates/feed.html")
	if err != nil {
		return nil, err
	}

	out := &bytes.Buffer{}
	err = tmpl.Execute(out, feed)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func formatTimeSincePost(timestamp time.Time) string {

	const (
		SecondsPerMinute = 60
		SecondsPerHour   = 60 * SecondsPerMinute
		SecondsPerDay    = 24 * SecondsPerHour
		SecondsPerMonth  = 30 * SecondsPerDay
		SecondsPerYear   = 365 * SecondsPerDay
	)

	pluralize := func(s string, val int) string {
		if val == 1 {
			return s
		}
		return s + "s"
	}

	seconds := int(time.Now().UTC().Sub(timestamp).Seconds())

	interval := seconds / SecondsPerYear
	if interval >= 1 {
		return fmt.Sprintf("%d %s", interval, pluralize("year", interval))
	}

	interval = seconds / SecondsPerMonth
	if interval >= 1 {
		return fmt.Sprintf("%d %s", interval, pluralize("month", interval))
	}

	interval = seconds / SecondsPerDay
	if interval >= 1 {
		return fmt.Sprintf("%d %s", interval, pluralize("day", interval))
	}

	interval = seconds / SecondsPerHour
	if interval >= 1 {
		return fmt.Sprintf("%d %s", interval, pluralize("hour", interval))
	}

	interval = seconds / SecondsPerMinute
	if interval >= 1 {
		return fmt.Sprintf("%d %s", interval, pluralize("minute", interval))
	}

	return fmt.Sprintf("%d %s", interval, pluralize("second", interval))
}

func typeString(t FeedPostType) string {
	return t.String()
}
