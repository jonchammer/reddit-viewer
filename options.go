package main

import (
	"errors"
	"fmt"
	"net/http"
)

// ------------------------------------------------------------------------- //
// Sort Method
// ------------------------------------------------------------------------- //

type SortMethod int

const (
	SortMethodDefault SortMethod = iota
	SortMethodHot
	SortMethodNew
	SortMethodRising
	SortMethodControversial
	SortMethodTop
	SortMethodGilded
)

func (sm SortMethod) URLString() string {
	switch sm {
	case SortMethodHot:
		return "hot"
	case SortMethodNew:
		return "new"
	case SortMethodRising:
		return "rising"
	case SortMethodControversial:
		return "controversial"
	case SortMethodTop:
		return "top"
	case SortMethodGilded:
		return "gilded"
	default:
		return ""
	}
}

func SortMethodFromString(s string) (SortMethod, error) {
	switch s {
	case "hot":
		return SortMethodHot, nil
	case "new":
		return SortMethodNew, nil
	case "rising":
		return SortMethodRising, nil
	case "controversial":
		return SortMethodControversial, nil
	case "top":
		return SortMethodTop, nil
	case "gilded":
		return SortMethodGilded, nil
	default:
		return SortMethod(-1), fmt.Errorf("'%s' is not a sort method", s)
	}
}

// ------------------------------------------------------------------------- //
// Feed Options
// ------------------------------------------------------------------------- //

type FeedOption func(opts *feedOpts) error

type feedOpts struct {

	// BaseURL will determine the host for the query. This will likely be
	// "old.reddit.com" when making external requests, but it might be empty
	// if referring to localhost.
	BaseURL string

	// Subreddit will be non-nil when accessing a particular subreddit (e.g.
	// /r/comics). When nil, data from the front page will be returned.
	Subreddit *string

	// SortMethod dictates how posts will be sorted. Must be one of the valid
	// members of the SortMethod enum.
	SortMethod SortMethod

	// Count is one method for paging. When provided it determines the index of
	// the first post to be returned. 0 == first page, 25 == second page,
	// 50 == third page, etc.
	//
	// Count can be set alongside LastPostID, but LastPostID will take
	// precedence if the two conflict with one another.
	Count int

	// LastPostID is one method for paging. If non-nil, it should be the ID of
	// the last feed post on the previous page. If nil, the first page is
	// returned.
	//
	// LastPostID can be set alongside Count, but LastPostID will take
	// precedence if the two conflict with one another.
	LastPostID *string

	// Any headers provided in the original HTTP request that should be
	// forwarded to Reddit.
	Headers http.Header
}

func WithBaseURL(baseURL string) FeedOption {
	return func(opts *feedOpts) error {
		opts.BaseURL = baseURL
		return nil
	}
}

func WithSubreddit(subreddit string) FeedOption {
	return func(opts *feedOpts) error {
		opts.Subreddit = &subreddit
		return nil
	}
}

func WithSortMethod(sortMethod SortMethod) FeedOption {
	return func(opts *feedOpts) error {
		if sortMethod < SortMethodDefault || sortMethod > SortMethodGilded {
			return errors.New("sort method not recognized")
		}

		opts.SortMethod = sortMethod
		return nil
	}
}

func WithCount(count int) FeedOption {
	return func(opts *feedOpts) error {
		opts.Count = count
		return nil
	}
}

func WithLastPostID(lastPostID string) FeedOption {
	return func(opts *feedOpts) error {
		opts.LastPostID = &lastPostID
		return nil
	}
}

func WithHeaders(headers http.Header) FeedOption {
	return func(opts *feedOpts) error {
		opts.Headers = headers
		return nil
	}
}
