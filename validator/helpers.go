// validator/helpers.go
package validator

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var filename string

// Errorf выводит ошибку в stderr.
// node может быть nil — тогда без номера строки.
func Errorf(node *yaml.Node, msg string) error {
	if node != nil && node.Line > 0 {
		fmt.Fprintf(os.Stderr, "%s:%d %s\n", filename, node.Line, msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", filename, msg)
	}
	return fmt.Errorf("%s", msg)
}

// findMappingNode ищет значение по ключу в mapping
func findMappingNode(parent *yaml.Node, key string) *yaml.Node {
	if parent == nil || parent.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(parent.Content); i += 2 {
		if i+1 < len(parent.Content) && parent.Content[i].Value == key {
			return parent.Content[i+1]
		}
	}
	return nil
}

// requireField — обязательное поле, при отсутствии — ошибка без номера строки
func requireField(parent *yaml.Node, field string) (*yaml.Node, error) {
	node := findMappingNode(parent, field)
	if node == nil {
		return nil, Errorf(nil, field + " is required")
	}
	return node, nil
}