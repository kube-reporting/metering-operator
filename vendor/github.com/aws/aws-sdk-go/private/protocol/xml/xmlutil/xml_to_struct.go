package xmlutil

import (
	"encoding/xml"
	"fmt"
	"io"
	"sort"
)

// A XMLNode contains the values to be encoded or decoded.
type XMLNode struct {
	Name     xml.Name              `json:",omitempty"`
	Children map[string][]*XMLNode `json:",omitempty"`
	Text     string                `json:",omitempty"`
	Attr     []xml.Attr            `json:",omitempty"`

	namespaces map[string]string
	parent     *XMLNode
}

// NewXMLElement returns a pointer to a new XMLNode initialized to default values.
func NewXMLElement(name xml.Name) *XMLNode {
	return &XMLNode{
		Name:     name,
		Children: map[string][]*XMLNode{},
		Attr:     []xml.Attr{},
	}
}

// AddChild adds child to the XMLNode.
func (n *XMLNode) AddChild(child *XMLNode) {
	if _, ok := n.Children[child.Name.Local]; !ok {
		n.Children[child.Name.Local] = []*XMLNode{}
	}
	n.Children[child.Name.Local] = append(n.Children[child.Name.Local], child)
}

// XMLToStruct converts a xml.Decoder stream to XMLNode with nested values.
func XMLToStruct(d *xml.Decoder, s *xml.StartElement) (*XMLNode, error) {
	out := &XMLNode{}
	for {
		tok, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			} ***REMOVED*** {
				return out, err
			}
		}

		if tok == nil {
			break
		}

		switch typed := tok.(type) {
		case xml.CharData:
			out.Text = string(typed.Copy())
		case xml.StartElement:
			el := typed.Copy()
			out.Attr = el.Attr
			if out.Children == nil {
				out.Children = map[string][]*XMLNode{}
			}

			name := typed.Name.Local
			slice := out.Children[name]
			if slice == nil {
				slice = []*XMLNode{}
			}
			node, e := XMLToStruct(d, &el)
			out.***REMOVED***ndNamespaces()
			if e != nil {
				return out, e
			}
			node.Name = typed.Name
			node.***REMOVED***ndNamespaces()
			tempOut := *out
			// Save into a temp variable, simply because out gets squashed during
			// loop iterations
			node.parent = &tempOut
			slice = append(slice, node)
			out.Children[name] = slice
		case xml.EndElement:
			if s != nil && s.Name.Local == typed.Name.Local { // matching end token
				return out, nil
			}
			out = &XMLNode{}
		}
	}
	return out, nil
}

func (n *XMLNode) ***REMOVED***ndNamespaces() {
	ns := map[string]string{}
	for _, a := range n.Attr {
		if a.Name.Space == "xmlns" {
			ns[a.Value] = a.Name.Local
		}
	}

	n.namespaces = ns
}

func (n *XMLNode) ***REMOVED***ndElem(name string) (string, bool) {
	for node := n; node != nil; node = node.parent {
		for _, a := range node.Attr {
			namespace := a.Name.Space
			if v, ok := node.namespaces[namespace]; ok {
				namespace = v
			}
			if name == fmt.Sprintf("%s:%s", namespace, a.Name.Local) {
				return a.Value, true
			}
		}
	}
	return "", false
}

// StructToXML writes an XMLNode to a xml.Encoder as tokens.
func StructToXML(e *xml.Encoder, node *XMLNode, sorted bool) error {
	e.EncodeToken(xml.StartElement{Name: node.Name, Attr: node.Attr})

	if node.Text != "" {
		e.EncodeToken(xml.CharData([]byte(node.Text)))
	} ***REMOVED*** if sorted {
		sortedNames := []string{}
		for k := range node.Children {
			sortedNames = append(sortedNames, k)
		}
		sort.Strings(sortedNames)

		for _, k := range sortedNames {
			for _, v := range node.Children[k] {
				StructToXML(e, v, sorted)
			}
		}
	} ***REMOVED*** {
		for _, c := range node.Children {
			for _, v := range c {
				StructToXML(e, v, sorted)
			}
		}
	}

	e.EncodeToken(xml.EndElement{Name: node.Name})
	return e.Flush()
}
