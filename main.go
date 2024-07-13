package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed static
var staticFiles embed.FS

func fileServer() http.Handler {
	return http.FileServer(http.FS(staticFiles))
}

func loggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logF(LevelDebug, "Got request: %s %s", r.Method, r.URL.String())
		h.ServeHTTP(w, r)
	})
}

type ProxyHandler struct {
	Parser *RedditParser
}

// ServeHTTP is the main request router for Reddit traffic. For feeds (front
// page and individual subreddits), we support:
//   - Sort Method (e.g. "hot", "top", etc.)
//   - Paging (e.g. "after=abcd")
//   - JSON output (if the URL ends with ".json")
//
// Front Page Routes:
//
//	[root]/
//	[root]/[sort_method]
//	[root]/[sort_method].json
//	[root]/?after=[last_post_id]
//	[root]/?after=[last_post_id].json
//	[root]/[sort_method]/?after=[last_post_id]
//	[root]/[sort_method]/?after=[last_post_id].json
//
// Subreddit Feed Routes:
//
//	[root]/r/foobar
//	[root]/r/foobar.json
//	[root]/r/foobar/[sort_method]
//	[root]/r/foobar/[sort_method].json
//	[root]/r/foobar/?after=[last_post_id]
//	[root]/r/foobar/?after=[last_post_id].json
//	[root]/r/foobar/[sort_method]/?after=[last_post_id]
//	[root]/r/foobar/[sort_method]/?after=[last_post_id].json
//
// Comments Routes:
//
//	[root]/r/foobar/comments/[post_id]/[some_title]/
//	[root]/r/foobar/comments/[post_id]/[some_title].json
func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Make sure we can recover gracefully from a panic
	defer func() {
		if e := recover(); e != nil {
			logF(LevelError, "Recovered from panic: %v", e)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	// Work out if the user intends for us to return JSON output or HTML
	outputJSON := false
	if strings.HasSuffix(r.URL.Path, ".json") {
		outputJSON = true
		r.URL.Path = strings.TrimSuffix(r.URL.Path, ".json")
	} else if strings.HasSuffix(r.URL.RawQuery, ".json") {
		outputJSON = true
		r.URL.RawQuery = strings.TrimSuffix(r.URL.RawQuery, ".json")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Invoke the parser to download the desired feed
	feed, err := ph.Parser.Feed(ctx, parseFeedOptions(r)...)
	if err != nil {
		logF(LevelError, "Failed to retrieve feed: %v", err)
		statusCode := http.StatusInternalServerError
		if httpErr, ok := err.(*HTTPError); ok {
			statusCode = httpErr.StatusCode
		}
		w.WriteHeader(statusCode)
		return
	}

	// Render as JSON
	if outputJSON {
		out, err := json.MarshalIndent(feed, "", "  ")
		if err != nil {
			logF(LevelError, "Failed to generate JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(out)
		return
	}

	// Render as HTML
	out, err := renderFeed(feed)
	if err != nil {
		logF(LevelError, "Failed to render feed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(out)
}

func parseFeedOptions(r *http.Request) []FeedOption {

	var options []FeedOption

	// Subreddit and sort method. Note that "/r/foobar/hot" is parsed as:
	// ["", "r", "foobar", "hot"] (with an empty first element).
	pieces := strings.Split(r.URL.Path, "/")
	if len(pieces) >= 3 && pieces[1] == "r" {
		options = append(options, WithSubreddit(pieces[2]))
	}
	if len(pieces) > 0 {
		if sm, err := SortMethodFromString(pieces[len(pieces)-1]); err == nil {
			options = append(options, WithSortMethod(sm))
		}
	}

	// Last Post ID
	if lastPostID := r.URL.Query().Get("after"); lastPostID != "" {
		options = append(options, WithLastPostID(lastPostID))
	}

	// Headers
	//   - NOTE: This parser doesn't currently handle 'gzip' or other
	//     compressed formats.
	headers := r.Header
	headers.Del("Accept-Encoding")
	options = append(options, WithHeaders(headers))

	return options
}

func main() {

	client, err := getDefaultHTTPClient()
	if err != nil {
		failF("failed to get default http client: %v", err)
	}

	server := &ProxyHandler{
		Parser: &RedditParser{
			Client: client,
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/favicon.ico", loggingHandler(http.NotFoundHandler()))
	mux.Handle("/static/", loggingHandler(fileServer()))
	mux.Handle("/", loggingHandler(server))

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		failF("failed to start server: %v", err)
	}
}

func failF(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
