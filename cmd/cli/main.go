package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Grady-Saccullo/schema-scrapper/pkg/ast"
	"github.com/chromedp/cdproto/emulation"
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
		emulation.SetDeviceMetricsOverride(
			3840,  // Width
			2160,  // Height
			2.0,   // Device scale factor
			false, // Mobile
		),
		emulation.SetUserAgentOverride("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15"),
		chromedp.Navigate(url),
		// chromedp.Sleep(10*time.Second),
		// chromedp.WaitVisible(`[data-e2e="explore-card-desc"]`, chromedp.ByQueryAll),
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

	nodeData, _ := json.MarshalIndent(nodes, "", "  ")

	err = os.WriteFile("data.json", nodeData, 0644)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	parseJsonSchema(nodes, nil, data, outdata)

	d, _ := json.MarshalIndent(outdata, "", "  ")

	// write to output file
	err = os.WriteFile("out.json", d, 0644)

	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

}

func getNode(tree ast.Node, lookUpNodes *map[string]ast.Node, path string) interface{} {
	s := strings.Split(path, ".")
	node := findNode(tree, lookUpNodes, s)

	if node == nil {
		log.Println("could not find node with path: ", path)
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

	return node
}

func findNode(node ast.Node, lookUpNodes *map[string]ast.Node, path []string) ast.Node {

	if len(path) == 0 {
		return node
	}

	switch n := node.(type) {
	case *ast.Element:
		possibleEl := path[0]
		// we know we need to look up a node

		if possibleEl[0] == '@' {
			if lookUpNodes == nil {
				panic("Error: $lookup_nodes is required for list")
			}

			possibleEl = possibleEl[1:]

			if _, ok := (*lookUpNodes)[possibleEl]; !ok {
				panic(fmt.Sprintf("Error: $lookup_nodes, %s, does not exist", possibleEl))
			}

			return findNode((*lookUpNodes)[possibleEl], lookUpNodes, path[1:])
		}

		tag, reservedAttrs, jsonAttr := getPathValueWithAttrs(path[0])
		nth := 0

		for _, child := range n.Children() {
			switch c := child.(type) {
			case *ast.Element:
				isTag := c.Tag() == ast.ElementTag(tag)
				hasAttrs := checkAttrs(c.Attributes(), jsonAttr)

				if isTag && hasAttrs {
					if reservedAttrs != nil && reservedAttrs.Nth != nil {
						if *reservedAttrs.Nth == nth {
							return findNode(child, lookUpNodes, path[1:])
						}
						nth++
					} else {
						return findNode(child, lookUpNodes, path[1:])
					}
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

func getPathValueWithAttrs(path string) (string, *ReservedAttrs, *map[string]string) {
	startIndex := strings.Index(path, "[")
	if startIndex == -1 {
		return path, nil, nil
	}

	endIndex := strings.Index(path, "]")

	if endIndex == -1 {
		panic("Invalid path format. Expected closing bracket, found: " + path)
	}

	tag := path[:startIndex]
	var out *map[string]string
	attrs := path[startIndex+1 : endIndex]
	attrParts := strings.Split(attrs, ",")

	if len(attrParts) == 0 {
		panic("Invalid attribute format. Expected key=value, found: " + attrs)
	}

	var reservedAttrs *ReservedAttrs

	for _, attr := range attrParts {
		p := strings.Split(attr, "=")
		if len(p) != 2 {
			panic("Invalid attribute format. Expected key=value, found: " + attr)
		}

		switch p[0] {
		case "!nth":
			nth, err := strconv.Atoi(p[1])
			if err != nil {
				panic("Invalid attribute format. Expected integer, found: " + p[1])
			}
			if reservedAttrs == nil {
				reservedAttrs = &ReservedAttrs{}
			}

			reservedAttrs.Nth = &nth
		default:
			if out == nil {
				m := make(map[string]string)
				out = &m
			}

			(*out)[p[0]] = p[1]
		}
	}

	if out == nil && reservedAttrs == nil {
		panic("Invalid attribute format. Expected key=value, found: " + attrs)
	}

	return tag, reservedAttrs, out
}

type ReservedAttrs struct {
	Nth *int
}

func parseJsonSchema(
	node ast.Node,
	lookUpNodes *map[string]ast.Node,
	jsonSchema map[string]interface{},
	out interface{},
) {
	for k, v := range jsonSchema {
		switch out := out.(type) {
		case map[string]interface{}:
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

					out[k] = getNode(node, lookUpNodes, el.Selector)
				case "number":
					fmt.Println("Not implemented yet")
				case "array":
					var el JsonSchemaNodeArray
					err := mapstructure.Decode(v, &el)

					if err != nil {
						fmt.Println("Error: ", err)
						return
					}

					if err = validateArrayJson(el); err != nil {
						fmt.Println("Error: ", err)
						return
					}

					itrElTree := getNode(node, lookUpNodes, *el.ItrEl)

					if itrElTree == nil {
						fmt.Println("Error: Could not find node for itr_el: ", *el.ItrEl)
						return
					}

					if el.ItrIdent == nil {
						panic("Error: $itr_ident is required for list")
					}

					if el.ValueNode == nil {
						panic("Error: $value_node is required for list")
					}

					out[k] = []interface{}{}

					idx := 0

					for _, childEl := range itrElTree.(*ast.Element).Children() {
						if el.StartOffset != nil && idx < *el.StartOffset {
							idx++
							continue
						}

						if lookUpNodes != nil {
							(*lookUpNodes)[*el.ItrIdent] = childEl
						} else {
							m := make(map[string]ast.Node)
							m[*el.ItrIdent] = childEl
							lookUpNodes = &m
						}

						outItem := make(map[string]interface{})

						parseJsonSchema(childEl, lookUpNodes, *el.ValueNode, outItem)

						out[k] = append(out[k].([]interface{}), outItem)

					}

					// parseJsonSchema(node, lookUpNodes, *el.ValueNode, out[k])

				case "object":
					var el JsonSchemaNodeObject
					err := mapstructure.Decode(v, &el)

					if err != nil {
						fmt.Println("Error: ", err)
						return
					}

					if err = validateObjectJson(el); err != nil {
						fmt.Println("Error: ", err)
						return
					}

					if el.ItrEl == nil {
						key := getNode(node, lookUpNodes, (*el.KeyNode)["$selector"].(string))
						if key == nil {
							panic("Error: Could not find node for key: " + (*el.KeyNode)["$selector"].(string))
						}

						switch (*el.ValueNode)["$type"] {
						case nil:
							value := map[string]interface{}{}
							parseJsonSchema(node, lookUpNodes, *el.ValueNode, value)
							out[k] = map[string]interface{}{key.(string): value}
						default:
							value := getNode(node, lookUpNodes, (*el.ValueNode)["$selector"].(string))
							out[k] = map[string]interface{}{key.(string): value}
						}
					} else {
						itrElTree := getNode(node, lookUpNodes, *el.ItrEl)

						if itrElTree == nil {
							fmt.Println("Error: Could not find node for itr_el: ", *el.ItrEl)
							return
						}

						if el.ItrIdent == nil {
							panic("Error: $itr_ident is required for object")
						}

						if el.ValueNode == nil {
							panic("Error: $value_node is required for object")
						}

						out[k] = make(map[string]interface{})

						idx := 0

						for _, childEl := range itrElTree.(*ast.Element).Children() {
							if el.StartOffset != nil && idx < *el.StartOffset {
								idx++
								continue
							}

							if lookUpNodes != nil {
								(*lookUpNodes)[*el.ItrIdent] = childEl
							} else {
								m := make(map[string]ast.Node)
								m[*el.ItrIdent] = childEl
								lookUpNodes = &m
							}

							switch (*el.ValueNode)["$type"] {
							case nil:
								outItem := map[string]interface{}{}
								parseJsonSchema(childEl, lookUpNodes, *el.ValueNode, outItem)

								key := getNode(childEl, lookUpNodes, (*el.KeyNode)["$selector"].(string))
								if key == nil {
									panic("Error: Could not find node for key: " + (*el.KeyNode)["$selector"].(string))
								}

								out[k].(map[string]interface{})[key.(string)] = outItem
							default:
								value := getNode(childEl, lookUpNodes, (*el.ValueNode)["$selector"].(string))
								key := getNode(childEl, lookUpNodes, (*el.KeyNode)["$selector"].(string))
								if key == nil {
									panic("Error: Could not find node for key: " + (*el.KeyNode)["$selector"].(string))
								}
								out[k].(map[string]interface{})[key.(string)] = value
							}
						}
					}
				case nil:
					out[k] = map[string]interface{}{}
					parseJsonSchema(node, lookUpNodes, v.(map[string]interface{}), out[k])
				default:
					fmt.Println("Unknown type: ", v.(map[string]interface{})["$type"])
				}
			case string:
				switch k {
				case "$version":
					fmt.Println("Version: ", v)
				case "$paths":
					fmt.Println("$paths not implemented yet... not too hook into look up nodes")
				default:
					panic(fmt.Sprintf("Error: %s:%s is not a valid schema", k, v))
				}
			}
		case []interface{}:
			switch v.(type) {
			case map[string]interface{}:
				parseJsonSchema(node, lookUpNodes, v.(map[string]interface{}), out)
			case string:
				fmt.Println("OUTER STRING: ", k, v)
			}
		}
	}
}

func validateArrayJson(el JsonSchemaNodeArray) error {
	if el.ItrEl == nil {
		return fmt.Errorf("Error: $itr_el is required for list")
	}

	if el.ValueNode == nil {
		return fmt.Errorf("Error: $value_node is required for list")

	}

	if el.ItrIdent == nil {
		return fmt.Errorf("Error: $itr_ident is required for list")
	}

	return nil
}

func validateObjectJson(el JsonSchemaNodeObject) error {
	if el.KeyNode == nil {
		return fmt.Errorf("Error: $key_node is required for object")
	}

	if el.ValueNode == nil {
		return fmt.Errorf("Error: $value_node is required for object")

	}

	if el.ItrEl != nil {
		if el.ItrIdent == nil {
			return fmt.Errorf("Error: $itr_ident is required for object when $itr_el is present")
		}
	}

	return nil
}

type SchemaNodeType string

const (
	SchemaNodeType_String SchemaNodeType = "string"
	SchemaNodeType_Number SchemaNodeType = "number"
	SchemaNodeType_Array  SchemaNodeType = "array"
	SchemaNodeType_Object SchemaNodeType = "object"
)

type JsonSchemaNodeString struct {
	Type     SchemaNodeType `json:"$type" mapstructure:"$type"`
	Selector string         `json:"$selector" mapstructure:"$selector"`
}

type JsonSchemaNodeNumber struct {
	Type     SchemaNodeType `json:"$type" mapstructure:"$type"`
	Selector string         `json:"$selector" mapstructure:"$selector"`
}

type JsonSchemaNodeArray struct {
	Type        SchemaNodeType          `json:"$type" mapstructure:"$type"`
	ValueNode   *map[string]interface{} `json:"$value_node" mapstructure:"$value_node"`
	ItrIdent    *string                 `json:"$itr_ident" mapstructure:"$itr_ident"`
	ItrEl       *string                 `json:"$itr_el" mapstructure:"$itr_el"`
	ItrElMatch  *string                 `json:"$itr_el_match" mapstructure:"$itr_el_match"`
	StartOffset *int                    `json:"$start_offset" mapstructure:"$start_offset"`
}

type JsonSchemaNodeObject struct {
	Type        SchemaNodeType          `json:"$type" mapstructure:"$type"`
	ItrIdent    *string                 `json:"$itr_ident" mapstructure:"$itr_ident"`
	ItrEl       *string                 `json:"$itr_el" mapstructure:"$itr_el"`
	ItrElMatch  *string                 `json:"$itr_el_match" mapstructure:"$itr_el_match"`
	KeyNode     *map[string]interface{} `json:"$key_node" mapstructure:"$key_node"`
	ValueNode   *map[string]interface{} `json:"$value_node" mapstructure:"$value_node"`
	StartOffset *int                    `json:"$start_offset" mapstructure:"$start_offset"`
}
