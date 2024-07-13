package main

import (
	"errors"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"regexp"
)

var (
	ErrSearchFailed = errors.New("failed to find HTML node with matching criteria")
)

func BreadthFirstSearch(
	root *html.Node,
	criteria SearchCriteria,
	recurseIf SearchCriteria,
) (*html.Node, error) {
	nodes := []*html.Node{
		root,
	}

	for {
		// Base case - we've exhausted our search
		if len(nodes) == 0 {
			break
		}

		// Extract the next candidate node from the front of the queue
		node := nodes[0]
		nodes = nodes[1:]

		// If we've found the node, return it.
		if criteria(node) {
			return node, nil
		}

		// Check if we should bother examining child elements. If so, those
		// will be added to the queue.
		if node.Type == html.DocumentNode || recurseIf(node) {
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				nodes = append(nodes, c)
			}
		}
	}

	return nil, ErrSearchFailed
}

func DepthFirstSearch(
	root *html.Node,
	criteria SearchCriteria,
	recurseIf SearchCriteria,
) (*html.Node, error) {
	nodes := []*html.Node{
		root,
	}

	for {
		// Base case - we've exhausted our search
		if len(nodes) == 0 {
			break
		}

		// Extract the next candidate node from the top of the stack
		node := nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]

		// If we've found the node, return it.
		if criteria(node) {
			return node, nil
		}

		// Check if we should bother examining child elements. If so, those
		// will be added to the stack as well.
		if node.Type == html.DocumentNode || recurseIf(node) {
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				nodes = append(nodes, c)
			}
		}
	}

	return nil, ErrSearchFailed
}

func NthChild(
	node *html.Node,
	n int,
) (*html.Node, error) {

	c := node.FirstChild
	if c == nil {
		return nil, errors.New("child index out of bounds")
	}

	for i := 0; i < n; i++ {
		c = c.NextSibling
		if c == nil {
			return nil, errors.New("child index out of bounds")
		}
	}

	return c, nil
}

func GetAttribute(node *html.Node, attributeKey string) (string, bool) {
	for _, attr := range node.Attr {
		if attr.Key == attributeKey {
			return attr.Val, true
		}
	}

	return "", false
}

// ------------------------------------------------------------------------- //
// Search Criteria
// ------------------------------------------------------------------------- //

type SearchCriteria func(*html.Node) bool

func And(criteria ...SearchCriteria) SearchCriteria {
	return func(node *html.Node) bool {
		for _, c := range criteria {
			if !c(node) {
				return false
			}
		}

		return true
	}
}

func Or(criteria ...SearchCriteria) SearchCriteria {
	return func(node *html.Node) bool {
		for _, c := range criteria {
			if c(node) {
				return true
			}
		}

		return false
	}
}

func Not(criteria SearchCriteria) SearchCriteria {
	return func(node *html.Node) bool {
		return !criteria(node)
	}
}

func IsTag(name atom.Atom) SearchCriteria {
	return func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.DataAtom == name
	}
}

func HasAttribute(attributeKey string) SearchCriteria {
	return func(node *html.Node) bool {
		for _, attr := range node.Attr {
			if attr.Key == attributeKey {
				return true
			}
		}
		return false
	}
}

func HasAttributeWithValue(attributeKey string, attributeValue string) SearchCriteria {
	return func(node *html.Node) bool {
		for _, attr := range node.Attr {
			if attr.Key == attributeKey && attr.Val == attributeValue {
				return true
			}
		}
		return false
	}
}

func HasAttributeWithValueRegex(attributeKey string, attributeRegex string) SearchCriteria {
	return func(node *html.Node) bool {
		for _, attr := range node.Attr {
			if attr.Key != attributeKey {
				continue
			}

			if matched, _ := regexp.MatchString(attributeRegex, attr.Val); matched {
				return true
			}
		}
		return false
	}
}

func RecurseAlways(_ *html.Node) bool {
	return true
}

func RecurseNever(_ *html.Node) bool {
	return false
}
