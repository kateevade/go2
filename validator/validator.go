// validator/validator.go
package validator

import (
	"os"

	"gopkg.in/yaml.v3"
)

func ValidatePodYAML(filepath string) error {
	filename = filepath

	content, err := os.ReadFile(filepath)
	if err != nil {
		return Errorf(nil, "cannot read file: %v", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return Errorf(&root, "cannot unmarshal yaml: %v", err)
	}

	if len(root.Content) == 0 {
		return Errorf(&root, "empty yaml document")
	}

	doc := root.Content[0]
	return validateTopLevel(doc)
}