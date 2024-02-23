package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Please provide a URL")
		return
	}

	jsonSchemaFilePath := args[1]

	url := args[0]

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	// need this to be eiter RootNode or ElementNode

	root := ElementNode{}
	currentNode := &root

	hasMore := true

	for hasMore {

		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			hasMore = false
		case html.StartTagToken:
			tagName, hasAttr := z.TagName()
			name := string(tagName)

			el := ElementNode{
				Tag:        AST_Node_Name(name),
				Type:       AST_Node_Type_Element,
				Attributes: map[string]string{},
				Parent:     currentNode,
			}

			if hasAttr {
				for {
					key, val, moreAttr := z.TagAttr()
					el.Attributes[string(key)] = string(val)
					if !moreAttr {
						break
					}
				}
			}

			currentNode.AddChild(&el)
			currentNode = &el

			if name == string(AST_Node_Name_Meta) {
				// if it's a meta tag, we don't need to parse the children
				currentNode = currentNode.GetParent()
			}

		case html.EndTagToken:
			if currentNode.Type == AST_Node_Type_Root {
				hasMore = false
			} else {
				currentNode = currentNode.GetParent()
			}
		case html.TextToken:
			text := strings.TrimSpace(string(z.Text()))
			if text == "" {
				continue
			}
			el := TextNode{
				Type:   AST_Node_Type_Text,
				Parent: currentNode,
				Value:  text,
			}
			currentNode.AddChild(&el)
		}
	}

	file, err := os.Open(jsonSchemaFilePath)

	defer file.Close()

	var data map[string]interface{}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	err = json.Unmarshal(fileBytes, &data)

	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	outdata := map[string]interface{}{}

	for k, v := range data {
		switch v.(type) {
		case string:
			// split the string by the dot
			s := strings.Split(v.(string), ".")
			node := findNode(root.GetChildren()[0], s)
			if node == nil {
				println("could not find node")
				continue
			}
			children := node.GetChildren()
			if len(children) > 0 {
				if children[0].GetType() == AST_Node_Type_Text {
					child := children[0].(*TextNode)
					if child.Value != "" {
						outdata[k] = child.Value
					}
				}
			}

		}
	}

	d, _ := json.MarshalIndent(outdata, "", "  ")
	fmt.Println(string(d))
}

func findNode(node Node, path []string) Node {
	if len(path) == 0 {
		return node
	}

	for _, child := range node.GetChildren() {
		if child.GetType() == AST_Node_Type_Element {
			el := child.(*ElementNode)
			if el.Tag == AST_Node_Name(path[0]) {
				return findNode(child, path[1:])
			}
		}
	}

	return nil
}

type AST_Node_Type string
type AST_Node_Name string

const (
	AST_Node_Type_Element AST_Node_Type = "element"
	AST_Node_Type_Text    AST_Node_Type = "text"
	AST_Node_Type_Comment AST_Node_Type = "comment"
	AST_Node_Type_Root    AST_Node_Type = "root"

	AST_Node_Name_Html       AST_Node_Name = "html"
	AST_Node_Name_Head       AST_Node_Name = "head"
	AST_Node_Name_Body       AST_Node_Name = "body"
	AST_Node_Name_Title      AST_Node_Name = "title"
	AST_Node_Name_Meta       AST_Node_Name = "meta"
	AST_Node_Name_Link       AST_Node_Name = "link"
	AST_Node_Name_Script     AST_Node_Name = "script"
	AST_Node_Name_Style      AST_Node_Name = "style"
	AST_Node_Name_P          AST_Node_Name = "p"
	AST_Node_Name_A          AST_Node_Name = "a"
	AST_Node_Name_Div        AST_Node_Name = "div"
	AST_Node_Name_Span       AST_Node_Name = "span"
	AST_Node_Name_H1         AST_Node_Name = "h1"
	AST_Node_Name_H2         AST_Node_Name = "h2"
	AST_Node_Name_H3         AST_Node_Name = "h3"
	AST_Node_Name_H4         AST_Node_Name = "h4"
	AST_Node_Name_H5         AST_Node_Name = "h5"
	AST_Node_Name_H6         AST_Node_Name = "h6"
	AST_Node_Name_Ul         AST_Node_Name = "ul"
	AST_Node_Name_Ol         AST_Node_Name = "ol"
	AST_Node_Name_Li         AST_Node_Name = "li"
	AST_Node_Name_Dl         AST_Node_Name = "dl"
	AST_Node_Name_Dt         AST_Node_Name = "dt"
	AST_Node_Name_Dd         AST_Node_Name = "dd"
	AST_Node_Name_Table      AST_Node_Name = "table"
	AST_Node_Name_Thead      AST_Node_Name = "thead"
	AST_Node_Name_Tbody      AST_Node_Name = "tbody"
	AST_Node_Name_Tfoot      AST_Node_Name = "tfoot"
	AST_Node_Name_Tr         AST_Node_Name = "tr"
	AST_Node_Name_Th         AST_Node_Name = "th"
	AST_Node_Name_Td         AST_Node_Name = "td"
	AST_Node_Name_Em         AST_Node_Name = "em"
	AST_Node_Name_Strong     AST_Node_Name = "strong"
	AST_Node_Name_B          AST_Node_Name = "b"
	AST_Node_Name_I          AST_Node_Name = "i"
	AST_Node_Name_U          AST_Node_Name = "u"
	AST_Node_Name_S          AST_Node_Name = "s"
	AST_Node_Name_Code       AST_Node_Name = "code"
	AST_Node_Name_Pre        AST_Node_Name = "pre"
	AST_Node_Name_Blockquote AST_Node_Name = "blockquote"
	AST_Node_Name_Hr         AST_Node_Name = "hr"
	AST_Node_Name_Br         AST_Node_Name = "br"
	AST_Node_Name_Img        AST_Node_Name = "img"
	AST_Node_Name_Input      AST_Node_Name = "input"
	AST_Node_Name_Textarea   AST_Node_Name = "textarea"
	AST_Node_Name_Select     AST_Node_Name = "select"
	AST_Node_Name_Option     AST_Node_Name = "option"
	AST_Node_Name_Form       AST_Node_Name = "form"
	AST_Node_Name_Fieldset   AST_Node_Name = "fieldset"
	AST_Node_Name_Legend     AST_Node_Name = "legend"
	AST_Node_Name_Label      AST_Node_Name = "label"
	AST_Node_Name_Button     AST_Node_Name = "button"
)

type Node interface {
	GetType() AST_Node_Type
	AddChild(Node)
	GetChildren() []Node
	SetParent(ElementNode)
	GetParent() *ElementNode
}

type ElementNode struct {
	Type       AST_Node_Type
	Tag        AST_Node_Name
	Attributes map[string]string
	Children   []Node
	Parent     *ElementNode `json:"-"`
}

func (e *ElementNode) AddChild(n Node) {
	e.Children = append(e.Children, n)
}

func (e *ElementNode) GetChildren() []Node {
	return e.Children
}

func (e *ElementNode) SetParent(n ElementNode) {
	e.Parent = &n
}

func (e *ElementNode) GetParent() *ElementNode {
	return e.Parent
}

func (e *ElementNode) GetType() AST_Node_Type {
	return e.Type
}

type TextNode struct {
	Type   AST_Node_Type
	Parent *ElementNode `json:"-"`
	Value  string
}

func (t *TextNode) AddChild(n Node) {}
func (t *TextNode) GetChildren() []Node {
	return []Node{}
}
func (t *TextNode) SetParent(n ElementNode) {
	t.Parent = &n
}

func (t *TextNode) GetParent() *ElementNode {
	return t.Parent
}

func (t *TextNode) GetType() AST_Node_Type {
	return t.Type
}

type CommentNode struct {
	Value string
}

func (c CommentNode) Type() AST_Node_Type {
	return AST_Node_Type_Comment
}
