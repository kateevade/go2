// validator/helpers.go
package validator

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var filename string

// Errorf выводит ошибку в stderr.
// Если node != nil и есть Line — выводит с номером строки.
// Если node == nil — только сообщение (для "is required").
func Errorf(node *yaml.Node, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)

	if node != nil && node.Line > 0 {
		fmt.Fprintf(os.Stderr, "%s:%d %s\n", filename, node.Line, msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", filename, msg)
	}
	return fmt.Errorf("%s", msg)
}

// findMappingNode ищет значение по ключу в mapping-узле
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

// requireField проверяет наличие обязательного поля.
// При отсутствии — ошибка БЕЗ номера строки (как требует задание).
func requireField(parent *yaml.Node, field string) (*yaml.Node, error) {
	node := findMappingNode(parent, field)
	if node == nil {
		return nil, Errorf(nil, "%s is required", field)
	}
	return node, nil
}