package ast

import (
	"io"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		r io.Reader
	}

	testReader := strings.NewReader(testhtml)

	tests := []struct {
		name    string
		args    args
		want    Node
		wantErr bool
	}{
		{
			name: "Testing nil reader",
			args: args{
				r: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Testing valid tree",
			args: args{
				r: testReader,
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.r)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewAst() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil {
			} else {
				if got != nil {
					t.Errorf("NewAst() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

var testhtml = `
<!DOCTYPE html>
<html>
	<head>
		<title>Page Title</title>
	</head>
	<body>
		<h1>This is a Heading</h1>
		<p>This is a paragraph.</p>
	</body>
</html>
`
// TODO: refactor / get working again
// var testHtmlAstNode = Element{
// 	NodeType: NodeTypeElement,
// 	ElTag:      ElementTag("html"),
// 	Attrs: &map[string]string{
// 		"lang": "en",
// 	},
// 	NodeChildren: []Node{
// 		&Element{
// 			NodeType: NodeTypeElement,
// 			ElTag:      ElementTag("head"),
// 			NodeChildren: []Node{
// 				&Element{
// 					NodeType: NodeTypeElement,
// 					ElTag:      ElementTag("title"),
// 					NodeChildren: []Node{
// 						&Text{
// 							nodeType: NodeTypeText,
// 							value:    "Page Title",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		&Element{
// 			NodeType: NodeTypeElement,
// 			ElTag:      ElementTag("body"),
// 			NodeChildren: []Node{
// 				&Element{
// 					NodeType: NodeTypeElement,
// 					ElTag:      ElementTag("h1"),
// 					NodeChildren: []Node{
// 						&Text{
// 							nodeType: NodeTypeText,
// 							value:    "This is a Heading",
// 						},
// 					},
// 				},
// 				&Element{
// 					NodeType: NodeTypeElement,
// 					ElTag:      ElementTag("p"),
// 					NodeChildren: []Node{
// 						&Text{
// 							nodeType: NodeTypeText,
// 							value:    "This is a paragraph.",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	},
// }
