package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Grady-Saccullo/schema-scrapper/pkg/ast"
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

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	defer resp.Body.Close()

	nodes, err := ast.New(resp.Body)
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

	tag, attrs := getPathValueWithAttrs(path[0])

	switch n := node.(type) {
	case *ast.Element:
		for _, child := range n.Children() {
			switch c := child.(type) {
			case *ast.Element:
				isTag := c.Tag() == ast.ElementTag(tag)
				hasAttrs := checkAttrs(*c.Attributes(), attrs)

				if isTag && hasAttrs {
					return findNode(child, path[1:])
				}
			}
		}
	}

	return nil
}

func checkAttrs(attrs map[string]string, nodeAttrs *map[string]string) bool {
	if nodeAttrs == nil {
		return true
	}

	for k, v := range *nodeAttrs {
		if attrs[k] != v {
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
				var el JsonSchemaString
				err := mapstructure.Decode(v, &el)
				if err != nil {
					fmt.Println("Error: ", err)
					return
				}

				println()
				out[k] = getNode(node, el.Node)
				println()
			case "list":
				fmt.Println("TODO: parse html into list")
				continue
				// var el JsonSchemaNodeList
				// err := mapstructure.Decode(v, &el)
				// if err != nil {
				// 	fmt.Println("Error: ", err)
				// 	return
				// }
				//
				// if el.Item != nil {
				// 	fmt.Println("parsing items")
				// 	out[k] = map[string]interface{}{}
				// 	parseJsonSchema(*el.Item, out[k].(map[string]interface{}))
				// }
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

type SchemaNodeType string

const (
	SchemaNodeType_String SchemaNodeType = "string"
	SchemaNodeType_List   SchemaNodeType = "list"
)

type JsonSchemaString struct {
	Type  SchemaNodeType `json:"$type" mapstructure:"$type"`
	Node  string         `json:"$node" mapstructure:"$node"`
	Index *int           `json:"$index" mapstructure:"$index"`
}

type JsonSchemaNodeList struct {
	Type      SchemaNodeType          `json:"$type" mapstructure:"$type"`
	Node      string                  `json:"$node" mapstructure:"$node"`
	ItrNode   string                  `json:"$itr_node" mapstructure:"$itr_node"`
	NodeMatch *string                 `json:"$itr_node_match" mapstructure:"$itr_node_match"`
	Item      *map[string]interface{} `json:"$item" mapstructure:"$item"`
}
