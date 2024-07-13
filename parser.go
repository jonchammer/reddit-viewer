package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type RedditParser struct {
	Client *http.Client
}

// Feed is used to access the front page or an individual subreddit.
func (rp *RedditParser) Feed(
	ctx context.Context,
	options ...FeedOption,
) (*Feed, error) {

	// Process user options
	opts := &feedOpts{
		BaseURL:    "http://old.reddit.com",
		Subreddit:  nil,
		SortMethod: SortMethodDefault,
		Count:      0,
		LastPostID: nil,
		Headers:    nil,
	}
	for _, opt := range options {
		err := opt(opts)
		if err != nil {
			return nil, err
		}
	}

	// Construct the URL
	getURL := constructURL(opts)
	logF(LevelTrace, "Issuing request: GET %s", getURL)

	// Make the proxy request, returning the full HTML tree
	doc, err := rp.getFeedDocument(ctx, getURL, opts.Headers)
	if err != nil {
		return nil, err
	}

	// Parse the feed from the HTML tree
	posts, err := getFeedPosts(doc)
	if err != nil {
		return nil, err
	}

	// Construct the next page link. Note that we want to direct the user back
	// to localhost, not to the main Reddit host.
	opts.BaseURL = ""
	opts.LastPostID = &posts[len(posts)-1].ID
	nextPageLink := constructURL(opts)

	return &Feed{
		Posts:        posts,
		NextPageLink: nextPageLink,
	}, nil
}

// ------------------------------------------------------------------------- //
// Helpers
// ------------------------------------------------------------------------- //

func constructURL(opts *feedOpts) string {

	getURL := opts.BaseURL
	if opts.Subreddit != nil {
		getURL = fmt.Sprintf("%s/r/%s", getURL, *opts.Subreddit)
	}
	values := url.Values{}
	if opts.SortMethod != SortMethodDefault {
		values.Set("sort", opts.SortMethod.URLString())
	}
	if opts.Count != 0 {
		values.Set("count", fmt.Sprintf("%d", opts.Count))
	}
	if opts.LastPostID != nil {
		values.Set("after", *opts.LastPostID)
	}
	if v := values.Encode(); v != "" {
		getURL = fmt.Sprintf("%s/?%s", getURL, v)
	}

	return getURL
}

func (rp *RedditParser) getFeedDocument(
	ctx context.Context,
	url string,
	headers http.Header,
) (*html.Node, error) {

	// headers := map[string]string{
	// 	"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	// }

	body, _, err := get(ctx, rp.Client, url, headers)
	if err != nil {
		return nil, err
	}

	return html.Parse(bytes.NewReader(body))
}

func getFeedPosts(doc *html.Node) ([]FeedPost, error) {

	// 1. Find the "siteTable" element
	siteTable, err := getSiteTable(doc)
	if err != nil {
		return nil, err
	}

	// 2. Iterate through the site table and extract each feed post that we find
	var posts []FeedPost
	for c := siteTable.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}

		if p, err := tryParseFeedPost(c); err == nil {
			posts = append(posts, *p)
		}
	}
	return posts, nil
}

func getSiteTable(n *html.Node) (*html.Node, error) {

	siteTable, err := BreadthFirstSearch(n,
		And(
			IsTag(atom.Div),
			HasAttributeWithValue("id", "siteTable"),
		),
		Not(IsTag(atom.Head)),
	)
	if err != nil {
		return nil, fmt.Errorf("site table not found: %w", err)
	}

	return siteTable, nil
}

// ------------------------------------------------------------------------- //
// FeedPost parser
// ------------------------------------------------------------------------- //

var (
	ErrNotAPost          = errors.New("not a post")
	ErrPostIsAd          = errors.New("post is an ad")
	ErrTitleNotFound     = errors.New("title not found")
	ErrThumbnailNotFound = errors.New("thumbnail not found")
	ErrCommentsNotFound  = errors.New("comments not found")
)

func tryParseFeedPost(n *html.Node) (*FeedPost, error) {

	// Gather what information we can from the parent node
	var id string
	var op string
	var subreddit string
	var timestamp time.Time
	var score int
	var commentCount int
	var postLink string
	var isSpoiler bool
	var isNSFW bool

	for _, attr := range n.Attr {

		switch attr.Key {

		// Padding elements that can be skipped
		case "class":
			if attr.Val == "clearleft" || attr.Val == "nav-buttons" {
				return nil, ErrNotAPost
			}

		// Regular FeedPost fields
		case "data-fullname":
			id = attr.Val
		case "data-author":
			op = attr.Val
		case "data-subreddit":
			subreddit = attr.Val
		case "data-timestamp":
			tsMillis, err := strconv.ParseInt(attr.Val, 10, 64)
			if err != nil {
				continue
			}
			timestamp = time.UnixMilli(tsMillis).UTC()
		case "data-score":
			s, err := strconv.Atoi(attr.Val)
			if err != nil {
				continue
			}
			score = s
		case "data-comments-count":
			cc, err := strconv.Atoi(attr.Val)
			if err != nil {
				continue
			}
			commentCount = cc
		case "data-url":
			postLink = attr.Val
		case "data-spoiler":
			sp, err := strconv.ParseBool(attr.Val)
			if err != nil {
				continue
			}
			isSpoiler = sp
		case "data-nsfw":
			nsfw, err := strconv.ParseBool(attr.Val)
			if err != nil {
				continue
			}
			isNSFW = nsfw

		// Known nodes that can be skipped
		case "data-adserver-impression-id":
			return nil, ErrPostIsAd
		}
	}

	// Try to find any remaining fields that are child elements
	title, _ := findTitle(n)
	thumbnailLink, _ := findThumbnailLink(n)
	commentsLink, _ := findCommentsLink(n)

	post := &FeedPost{
		ID:            id,
		Type:          FeedPostTypeLink,
		Title:         title,
		OP:            op,
		Subreddit:     subreddit,
		Timestamp:     timestamp,
		Score:         score,
		CommentCount:  commentCount,
		ThumbnailLink: thumbnailLink,
		PostLink:      postLink,
		CommentsLink:  commentsLink,
		IsSpoiler:     isSpoiler,
		IsNSFW:        isNSFW,
	}
	post.Type = classifyFeedPost(post)
	return post, nil
}

func findTitle(n *html.Node) (string, error) {

	// <div class="entry ...">
	//   <div class="top-matter">
	// 	   <p class="title">
	// 	     <a class="title ...">[TITLE]</a>
	// 	       ...
	// 	   </p>
	// 	   ...
	//   </div>
	//   ...
	// </div>
	titleNode, err := DepthFirstSearch(n,
		And(
			IsTag(atom.A),
			HasAttributeWithValueRegex("class", "title.*"),
		),
		RecurseAlways,
	)
	if err != nil {
		return "", ErrTitleNotFound
	}

	return titleNode.FirstChild.Data, nil
}

func findThumbnailLink(n *html.Node) (string, error) {

	// One of the child elements for 'n' is expected to have an <a class="thumbnail ...">
	// child that represents the visible thumbnail. This element should exist
	// even if there is no actual thumbnail, as Reddit renders a placeholder
	// icon if one isn't actually available.
	thumbnailNode, err := BreadthFirstSearch(n,
		And(
			IsTag(atom.A),
			HasAttributeWithValueRegex("class", "thumbnail.*"),
		),
		RecurseAlways,
	)
	if err != nil {
		return "", ErrThumbnailNotFound
	}

	// The thumbnail node may or may not have an <img> child tag. If it does,
	// we'll use that as the thumbnail link. If not, Reddit will render a
	// placeholder image, and we'll return failure.
	imgNode, err := BreadthFirstSearch(thumbnailNode,
		And(
			IsTag(atom.Img),
			HasAttribute("src"),
		),
		RecurseAlways,
	)
	if err != nil {
		return "", ErrThumbnailNotFound
	}

	src, _ := GetAttribute(imgNode, "src")
	return "https://" + strings.TrimPrefix(src, "//"), nil
}

func findCommentsLink(n *html.Node) (string, error) {
	// <div class="entry ...">
	//   <div class="top-matter">
	//     ...
	//     <ul class="flat-list buttons">
	//       <li class="first">
	//         <a href="[COMMENTS LINK]"/>
	//       </li>
	//       ...
	//     </ul>
	//   </div>
	//   ...
	// </div>

	// We expect to find a <ul> that represents the horizontal bar containing
	// some links/buttons (including the comments)
	buttonsBar, err := BreadthFirstSearch(n,
		IsTag(atom.Ul),
		RecurseAlways,
	)
	if err != nil {
		return "", ErrCommentsNotFound
	}

	// There should be a <li class="first"> element in the buttons bar that
	// will be the parent of the comments link.
	commentsNode, err := BreadthFirstSearch(buttonsBar,
		And(
			IsTag(atom.Li),
			HasAttributeWithValue("class", "first"),
		),
		RecurseAlways,
	)
	if err != nil {
		return "", ErrCommentsNotFound
	}

	// The comments link will be in an <a href="[link]" class="... comments ...">
	// element. The 'href' is the part we're looking for.
	commentsLink, err := BreadthFirstSearch(commentsNode,
		And(
			IsTag(atom.A),
			HasAttributeWithValueRegex("class", ".*comments.*"),
		),
		RecurseAlways,
	)
	if err != nil {
		return "", ErrCommentsNotFound
	}

	if link, ok := GetAttribute(commentsLink, "href"); ok {
		return link, nil
	}
	return "", ErrCommentsNotFound
}

func classifyFeedPost(post *FeedPost) FeedPostType {
	if strings.HasPrefix(post.PostLink, "/r/") {
		return FeedPostTypeText
	}
	if strings.HasSuffix(post.PostLink, ".jpg") ||
		strings.HasSuffix(post.PostLink, ".png") {
		return FeedPostTypeImage
	}
	if strings.HasPrefix(post.PostLink, "https://www.reddit.com/gallery/") {
		return FeedPostTypeGallery
	}
	if strings.HasPrefix(post.PostLink, "https://v.reddit.com/") ||
		strings.HasPrefix(post.PostLink, "https://v.redd.it/") {
		return FeedPostTypeVideo
	}

	return FeedPostTypeLink
}
