// validator/helpers.go
package validator

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var filename string

func Errorf(node *yaml.Node, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)

	if node != nil && node.Line > 0 {
		fmt.Fprintf(os.Stderr, "%s:%d %s\n", filename, node.Line, msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", filename, msg)
	}
	return fmt.Errorf(msg)
}

func findMappingNode(parent *yaml.Node, key string) *yaml.Node {
	if parent.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			if i+1 < len(parent.Content) {
				return parent.Content[i+1]
			}
			return nil // ключ есть, но значения нет (null)
		}
	}
	return nil
}

func hasKey(parent *yaml.Node, key string) bool {
	if parent.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			return true
		}
	}
	return false
}

func requireField(parent *yaml.Node, field string) (*yaml.Node, error) {
	node := findMappingNode(parent, field)
	if node == nil && !hasKey(parent, field) {
		return nil, Errorf(nil, "%s is required", field)
	}
	return node, nil
}