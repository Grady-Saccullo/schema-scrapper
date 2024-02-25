package ast

import (
	"fmt"
	"io"

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
			startTagToken(z, root)
		case html.EndTagToken:
			endTagToken(z, root)
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

func startTagToken(z *html.Tokenizer, currentNode Node) {
	tagName, hasAttr := z.TagName()
	name := string(tagName)

	el := Element{
		tag:      ElementTag(name),
		nodeType: NodeTypeElement,
		attrs:    nil,
		parent:   &currentNode,
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

	switch n := currentNode.(type) {
	case *Element:
		n.AddChild(&el)
		currentNode = &el
		// if the tag is a single tag, we need to go back to the parent
		if _, ok := singleTags[n.tag]; ok {
			currentNode = *currentNode.Parent()
		}
	case *Text:
		panic("text node cannot have children")
	}
}

func endTagToken(z *html.Tokenizer, currentNode Node) {
	parent := currentNode.Parent()
	if parent != nil {
		currentNode = *parent
	}
}

func textToken(z *html.Tokenizer, currentNode Node) {
	text := string(z.Text())
	textNode := Text{
		nodeType: NodeTypeText,
		parent:   &currentNode,
		value:    text,
	}

	switch n := currentNode.(type) {
	case *Element:
		n.AddChild(&textNode)
	case *Text:
		panic("text node cannot have children")
	}
}

var singleTags = map[ElementTag]interface{}{
	NodeTagArea:   nil,
	NodeTagBase:   nil,
	NodeTagBr:     nil,
	NodeTagCol:    nil,
	NodeTagHr:     nil,
	NodeTagImg:    nil,
	NodeTagInput:  nil,
	NodeTagLink:   nil,
	NodeTagMeta:   nil,
	NodeTagParam:  nil,
	NodeTagSource: nil,
	NodeTagTrack:  nil,
	NodeTagWbr:    nil,
}
