package ast

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

func New(r io.Reader) (Node, error) {
	if r == nil {
		return nil, fmt.Errorf("reader is nil")
	}

	z := html.NewTokenizer(r)

	root := &Element{}

	running := true

	for running {
		switch z.Next() {
		case html.StartTagToken:
			if node := startTagToken(z, root); node != nil {
				root = node.(*Element)
			}
		case html.EndTagToken:
			if node := endTagToken(z, root); node != nil {
				root = node.(*Element)
			}
		case html.TextToken:
			textToken(z, root)
		case html.CommentToken: // ignore comments for now
		case html.DoctypeToken: // ignore doctype for now
		case html.ErrorToken:
			err := z.Err()
			if err == io.EOF {
				running = false
			} else {
				return nil, err
			}
		}

	}

	// remove the empty root node
	return root.Children()[0], nil
}

func startTagToken(z *html.Tokenizer, node Node) Node {
	switch n := node.(type) {
	case *Element:
		tagName, hasAttr := z.TagName()
		name := string(tagName)

		el := Element{
			Name:       ElementTag(name),
			NodeType:   NodeTypeElement,
			Attrs:      nil,
			NodeParent: &node,
		}

		if hasAttr {
			for {
				key, val, moreAttr := z.TagAttr()
				el.SetAttribute(string(key), string(val))
				if !moreAttr {
					break
				}
			}
		}

		n.AddChild(&el)

		if _, ok := singleTags[el.Name]; ok {
			return n
		} else {
			return &el
		}
	case *Text:
		panic("text node cannot have children")
	default:
		err := fmt.Errorf("unknown node type: %T", n)
		panic(err)
	}
}

func endTagToken(z *html.Tokenizer, currentNode Node) Node {
	parent := currentNode.Parent()
	if parent != nil {
		return *parent
	}

	return nil
}

func textToken(z *html.Tokenizer, node Node) {
	text := strings.TrimSpace(string(z.Text()))
	if text == "" {
		return
	}

	switch n := node.(type) {
	case *Element:
		el := Text{
			NodeType:   NodeTypeText,
			NodeParent: &node,
			NodeValue:  text,
		}

		n.AddChild(&el)
	case *Text:
		panic("text node cannot have children")
	}
}

var singleTags = map[ElementTag]ElementTag{
	NodeTagArea:   NodeTagArea,
	NodeTagBase:   NodeTagBase,
	NodeTagBr:     NodeTagBr,
	NodeTagCol:    NodeTagCol,
	NodeTagHr:     NodeTagHr,
	NodeTagImg:    NodeTagImg,
	NodeTagInput:  NodeTagInput,
	NodeTagLink:   NodeTagLink,
	NodeTagMeta:   NodeTagMeta,
	NodeTagParam:  NodeTagParam,
	NodeTagSource: NodeTagSource,
	NodeTagTrack:  NodeTagTrack,
	NodeTagWbr:    NodeTagWbr,
}
