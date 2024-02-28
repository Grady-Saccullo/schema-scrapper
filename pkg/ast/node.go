package ast

type Node interface {
	Type() NodeType
	Parent() *Node
	SetParent(n *Node)
}

type NodeElement interface {
	Node
	Tag() ElementTag
	AddChild(n Node)
	Children() []Node
	SetAttribute(key, value string)
	Attributes() *map[string]string
}

type NodeText interface {
	Node
	Value() string
}

type Element struct {
	NodeType     NodeType           `json:"type"`
	Name         ElementTag         `json:"tag"`
	Attrs        *map[string]string `json:"attrs"`
	NodeChildren []Node             `json:"children"`
	NodeParent   *Node              `json:"-"` // for testing purposes (serialization)
}

func (e *Element) Type() NodeType {
	return e.NodeType
}

func (e *Element) Tag() ElementTag {
	return e.Name
}

func (e *Element) AddChild(n Node) {
	e.NodeChildren = append(e.NodeChildren, n)
}

func (e *Element) Children() []Node {
	return e.NodeChildren
}

func (e *Element) SetParent(n *Node) {
	e.NodeParent = n
}

func (e *Element) Parent() *Node {
	return e.NodeParent
}

func (e *Element) SetAttribute(key, value string) {
	if e.Attrs == nil {
		e.Attrs = &map[string]string{}
	}

	(*e.Attrs)[key] = value
}

func (e *Element) Attributes() *map[string]string {
	return e.Attrs
}

type Text struct {
	NodeType   NodeType `json:"type"`
	NodeParent *Node    `json:"-"` // for testing purposes (serialization)
	NodeValue  string   `json:"value"`
}

func (t *Text) SetParent(n *Node) {
	t.NodeParent = n
}

func (t *Text) Parent() *Node {
	return t.NodeParent
}

func (t *Text) Type() NodeType {
	return t.NodeType
}

func (t *Text) Value() string {
	return t.NodeValue
}
