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
	nodeType NodeType
	tag      ElementTag
	attrs    *map[string]string
	children []Node
	parent   *Node `json:"-"` // for testing purposes (serialization)
}

func (e *Element) Type() NodeType {
	return e.nodeType
}

func (e *Element) Tag() ElementTag {
	return e.tag
}

func (e *Element) AddChild(n Node) {
	e.children = append(e.children, n)
}

func (e *Element) Children() []Node {
	return e.children
}

func (e *Element) SetParent(n *Node) {
	e.parent = n
}

func (e *Element) Parent() *Node {
	return e.parent
}

func (e *Element) SetAttribute(key, value string) {
	if e.attrs == nil {
		e.attrs = &map[string]string{}
	}

	(*e.attrs)[key] = value
}

func (e *Element) Attributes() *map[string]string {
	return e.attrs
}

type Text struct {
	nodeType NodeType
	parent   *Node `json:"-"` // for testing purposes (serialization)
	value    string
}

func (t *Text) SetParent(n *Node) {
	t.parent = n
}

func (t *Text) Parent() *Node {
	return t.parent
}

func (t *Text) Type() NodeType {
	return t.nodeType
}

func (t *Text) Value() string {
	return t.value
}
