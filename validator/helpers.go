// validator/helpers.go
package validator

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Errorf — выводит ошибку в stderr в требуемом формате
func Errorf(node *yaml.Node, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	if node != nil && node.Line > 0 {
		fmt.Fprintf(os.Stderr, "%s:%d %s\n", filename, node.Line, msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", filename, msg)
	}
	return fmt.Errorf(msg)
}

var filename string // будет установлен в ValidatePodYAML

// findMappingNode ищет узел с заданным ключом в mapping-узле
func findMappingNode(parent *yaml.Node, key string) *yaml.Node {
	if parent.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(parent.Content); i += 2 {
		k := parent.Content[i]
		if k.Value == key {
			return parent.Content[i+1]
		}
	}
	return nil
}

// requireField проверяет наличие обязательного поля
func requireField(parent *yaml.Node, field string) (*yaml.Node, error) {
	node := findMappingNode(parent, field)
	if node == nil {
		keyNode := findKeyNode(parent, field)
		return nil, Errorf(keyNode, "%s is required", field)
	}
	return node, nil
}

// findKeyNode находит узел-ключа (для ошибки на правильной строке)
func findKeyNode(parent *yaml.Node, key string) *yaml.Node {
	if parent.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(parent.Content); i += 2 {
		k := parent.Content[i]
		if k.Value == key {
			return k
		}
	}
	return parent // если не нашли — возвращаем родителя
}