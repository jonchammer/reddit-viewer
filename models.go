package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Feed struct {
	Posts        []FeedPost `json:"posts"`
	NextPageLink string     `json:"nextPageLink"`
}

type FeedPost struct {
	ID            string       `json:"id"`
	Type          FeedPostType `json:"type"`
	Title         string       `json:"title"`
	OP            string       `json:"op"`
	Subreddit     string       `json:"subreddit"`
	Timestamp     time.Time    `json:"timestamp"`
	Score         int          `json:"score"`
	CommentCount  int          `json:"commentCount"`
	ThumbnailLink string       `json:"thumbnailLink"`
	PostLink      string       `json:"postLink"`
	CommentsLink  string       `json:"commentsLink"`
	IsSpoiler     bool         `json:"isSpoiler"`
	IsNSFW        bool         `json:"isNSFW"`
}

// ------------------------------------------------------------------------- //
// FeedPostType
// ------------------------------------------------------------------------- //

type FeedPostType int

const (
	FeedPostTypeText FeedPostType = iota
	FeedPostTypeLink
	FeedPostTypeImage
	FeedPostTypeVideo
	FeedPostTypeGallery
)

func (f FeedPostType) String() string {
	switch f {
	case FeedPostTypeText:
		return "text"
	case FeedPostTypeLink:
		return "link"
	case FeedPostTypeImage:
		return "image"
	case FeedPostTypeVideo:
		return "video"
	case FeedPostTypeGallery:
		return "gallery"
	default:
		return fmt.Sprintf("FeedPostType(%d)", f)
	}
}

func (f FeedPostType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, f.String())), nil
}

func (f *FeedPostType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("FeedPostType should be a string, got %s", data)
	}

	var t FeedPostType
	switch s {
	case "text":
		t = FeedPostTypeText
	case "link":
		t = FeedPostTypeLink
	case "image":
		t = FeedPostTypeImage
	case "video":
		t = FeedPostTypeVideo
	case "gallery":
		t = FeedPostTypeGallery
	default:
		return fmt.Errorf("%s does not belong to FeedPostType values", s)
	}

	*f = t
	return nil
}
