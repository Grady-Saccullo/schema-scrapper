package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Grady-Saccullo/schema-scrapper/pkg/ast"
	"github.com/chromedp/chromedp"
	"github.com/mitchellh/mapstructure"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Please provide a URL")
		return
	}

	jsonSchemaFilePath := args[1]

	url := args[0]

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &htmlContent),
	)

	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	contentIoReader := strings.NewReader(htmlContent)

	nodes, err := ast.New(contentIoReader)

	if err != nil {
		fmt.Println("Error: ", err)
		return
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

	d, _ := json.MarshalIndent(nodes, "", "  ")

	// write to output file
	err = os.WriteFile("out.json", d, 0644)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	parseJsonSchema(nodes, data, outdata)

	out, _ := json.MarshalIndent(outdata, "", "  ")
	fmt.Println(string(out))

}

func getNode(tree ast.Node, path string) interface{} {
	s := strings.Split(path, ".")
	node := findNode(tree, s)
	if node == nil {
		println("could not find node")
		return nil
	}

	switch n := node.(type) {
	case *ast.Element:
		children := n.Children()
		if len(children) > 0 {
			switch c := children[0].(type) {
			case *ast.Text:
				return c.Value()
			}
		}
	}

	return nil
}

func findNode(node ast.Node, path []string) ast.Node {

	if len(path) == 0 {
		return node
	}

	switch n := node.(type) {
	case *ast.Element:
		tag, jsonAttr := getPathValueWithAttrs(path[0])
		for _, child := range n.Children() {
			switch c := child.(type) {
			case *ast.Element:
				isTag := c.Tag() == ast.ElementTag(tag)
				hasAttrs := checkAttrs(c.Attributes(), jsonAttr)

				if isTag && hasAttrs {
					return findNode(child, path[1:])
				}
			}
		}
	}

	return nil
}

func checkAttrs(astAttrs *map[string]string, jsonAttrs *map[string]string) bool {
	if jsonAttrs == nil {
		return true
	}

	if astAttrs == nil {
		return false
	}

	for k, v := range *jsonAttrs {
		if (*astAttrs)[k] != v {
			return false
		}
	}

	return true
}

func getPathValueWithAttrs(path string) (string, *map[string]string) {
	out := map[string]string{}

	parts := strings.Split(path, "[")
	if len(parts) == 1 {
		return path, nil
	}

	path = parts[0]
	attrs := strings.Split(parts[1], "]")[0]
	attrParts := strings.Split(attrs, ",")
	for _, attr := range attrParts {
		p := strings.Split(attr, "=")
		out[p[0]] = p[1]
	}

	return path, &out
}

func parseJsonSchema(node ast.Node, jsonSchema map[string]interface{}, out map[string]interface{}) {
	for k, v := range jsonSchema {
		switch v.(type) {
		case map[string]interface{}:
			switch v.(map[string]interface{})["$type"] {
			case "string":
				var el JsonSchemaNodeString
				err := mapstructure.Decode(v, &el)
				if err != nil {
					fmt.Println("Error: ", err)
					return
				}

				out[k] = getNode(node, el.Node)
			case "list":
				var el JsonSchemaNodeList
				err := mapstructure.Decode(v, &el)

				if err != nil {
					fmt.Println("Error: ", err)
					return
				}

				if err = validateListJson(el); err != nil {
					fmt.Println("Error: ", err)
					return
				}

				nestedNode := getNode(node, *el.ItrNode)

				if el.Item != nil {
					out[k] = map[string]interface{}{}
					parseJsonSchema(node, *el.Item, out[k].(map[string]interface{}))
				}
			case nil:
				out[k] = map[string]interface{}{}
				parseJsonSchema(node, v.(map[string]interface{}), out[k].(map[string]interface{}))
			default:
				fmt.Println("Unknown type: ", v.(map[string]interface{})["$type"])
			}
		case string:
			fmt.Println("STRING: ", k, v)
		}
	}
}

func validateListJson(el JsonSchemaNodeList) error {
	if el.ItrNode == nil {
		return fmt.Errorf("Error: $itr_node is required for list")
	}

	if el.Item == nil {
		return fmt.Errorf("Error: $item is required for list")

	}

	if el.ItrIdent == nil {
		return fmt.Errorf("Error: $itr_ident is required for list")
	}

	return nil
}

type SchemaNodeType string

const (
	SchemaNodeType_String SchemaNodeType = "string"
	SchemaNodeType_Number SchemaNodeType = "number"
	SchemaNodeType_List   SchemaNodeType = "list"
	SchemaNodeType_Map    SchemaNodeType = "map"
)

type JsonSchemaNodeString struct {
	Type  SchemaNodeType `json:"$type" mapstructure:"$type"`
	Node  string         `json:"$node" mapstructure:"$node"`
	Index *int           `json:"$index" mapstructure:"$index"`
}

type JsonSchemaNodeNumber struct {
	Type  SchemaNodeType `json:"$type" mapstructure:"$type"`
	Node  string         `json:"$node" mapstructure:"$node"`
	Index *int           `json:"$index" mapstructure:"$index"`
}

type JsonSchemaNodeList struct {
	Item         *map[string]interface{} `json:"$item" mapstructure:"$item"`
	ItrIdent     *string                 `json:"$itr_ident" mapstructure:"$itr_ident"`
	ItrNode      *string                 `json:"$itr_node" mapstructure:"$itr_node"`
	ItrNodeMatch *string                 `json:"$itr_node_match" mapstructure:"$itr_node_match"`
	StartOffset  *int                    `json:"$start_offset" mapstructure:"$start_offset"`
	Type         SchemaNodeType          `json:"$type" mapstructure:"$type"`
}

type JsonSchemaNodeMap struct {
	Type SchemaNodeType `json:"$type" mapstructure:"$type"`
	// KeyNode must be of type string or number
	KeyNode      interface{} `json:"$key_node" mapstructure:"$key_node"`
	ValueNode    interface{} `json:"$value_node" mapstructure:"$value_node"`
	ItrNode      *string     `json:"$itr_node" mapstructure:"$itr_node"`
	ItrIdent     *string     `json:"$itr_ident" mapstructure:"$itr_ident"`
	ItrNodeMatch *string     `json:"$itr_node_match" mapstructure:"$itr_node_match"`
}
