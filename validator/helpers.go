// validator/helpers.go
package validator

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var filename string // устанавливается в ValidatePodYAML

// Errorf выводит ошибку в stderr в требуемом формате и возвращает error
func Errorf(node *yaml.Node, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	if node != nil && node.Line > 0 {
		fmt.Fprintf(os.Stderr, "%s:%d %s\n", filename, node.Line, msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", filename, msg)
	}
	return fmt.Errorf(msg)
}

// findMappingNode ищет дочерний узел по ключу в mapping
func findMappingNode(parent *yaml.Node, key string) *yaml.Node {
	if parent.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			return parent.Content[i+1]
		}
	}
	return nil
}

// findKeyNode ищет узел ключа (для правильной строки ошибки при отсутствии поля)
func findKeyNode(parent *yaml.Node, key string) *yaml.Node {
	if parent.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			return parent.Content[i]
		}
	}
	return parent // если не нашли — возвращаем родителя
}

// requireField проверяет наличие обязательного поля и возвращает его узел
func requireField(parent *yaml.Node, field string) (*yaml.Node, error) {
	node := findMappingNode(parent, field)
	if node == nil {
		return nil, Errorf(findKeyNode(parent, field), "%s is required", field)
	}
	return node, nil
}