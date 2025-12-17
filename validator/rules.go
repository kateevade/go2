// validator/rules.go
package validator

import (
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func validateTopLevel(root *yaml.Node) error {
	if root.Kind != yaml.MappingNode {
		return Errorf(root, "root must be a mapping")
	}

	apiVersion, err := requireField(root, "apiVersion")
	if err != nil {
		return err
	}
	if apiVersion.Kind != yaml.ScalarNode || apiVersion.Value != "v1" {
		return Errorf(apiVersion, "apiVersion has unsupported value '%s'", apiVersion.Value)
	}

	kind, err := requireField(root, "kind")
	if err != nil {
		return err
	}
	if kind.Kind != yaml.ScalarNode || kind.Value != "Pod" {
		return Errorf(kind, "kind has unsupported value '%s'", kind.Value)
	}

	metadata, err := requireField(root, "metadata")
	if err != nil {
		return err
	}
	if err := validateMetadata(metadata); err != nil {
		return err
	}

	spec, err := requireField(root, "spec")
	if err != nil {
		return err
	}
	return validateSpec(spec)
}

func validateMetadata(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return Errorf(node, "metadata must be object")
	}

	nameNode, err := requireField(node, "name")
	if err != nil {
		return err
	}
	if nameNode.Kind != yaml.ScalarNode {
		return Errorf(nameNode, "metadata.name must be string")
	}
	// Разрешаем пустое имя

	return nil
}

func validateSpec(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return Errorf(node, "spec must be object")
	}

	if osNode := findMappingNode(node, "os"); osNode != nil {
		if err := validatePodOS(osNode); err != nil {
			return err
		}
	}

	containersNode, err := requireField(node, "containers")
	if err != nil {
		return err
	}
	if containersNode.Kind != yaml.SequenceNode {
		return Errorf(containersNode, "spec.containers must be array")
	}

	for _, contNode := range containersNode.Content {
		if err := validateContainer(contNode); err != nil {
			return err
		}
	}

	return nil
}

func validatePodOS(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return Errorf(node, "spec.os must be object")
	}

	nameNode, err := requireField(node, "name")
	if err != nil {
		return err
	}
	if nameNode.Kind != yaml.ScalarNode {
		return Errorf(nameNode, "spec.os.name must be string")
	}
	// Любое значение

	return nil
}

func validateContainer(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return Errorf(node, "container must be object")
	}

	nameNode, err := requireField(node, "name")
	if err != nil {
		return err
	}
	if nameNode.Kind != yaml.ScalarNode {
		return Errorf(nameNode, "containers.name must be string")
	}
	// Разрешаем пустое имя контейнера

	imageNode, err := requireField(node, "image")
	if err != nil {
		return err
	}
	if imageNode.Kind != yaml.ScalarNode {
		return Errorf(imageNode, "containers.image must be string")
	}
	image := imageNode.Value
	if !strings.HasPrefix(image, "registry.bigbrother.io/") || !strings.Contains(image, ":") {
		return Errorf(imageNode, "containers.image has invalid format '%s'", image)
	}

	if portsNode := findMappingNode(node, "ports"); portsNode != nil {
		if portsNode.Kind != yaml.SequenceNode {
			return Errorf(portsNode, "containers.ports must be array")
		}
		for _, portNode := range portsNode.Content {
			if err := validateContainerPort(portNode); err != nil {
				return err
			}
		}
	}

	for _, probeName := range []string{"readinessProbe", "livenessProbe"} {
		if probeNode := findMappingNode(node, probeName); probeNode != nil {
			if err := validateProbe(probeNode); err != nil {
				return err
			}
		}
	}

	resourcesNode, err := requireField(node, "resources")
	if err != nil {
		return err
	}
	return validateResources(resourcesNode)
}

func validateContainerPort(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return Errorf(node, "containerPort item must be object")
	}

	portNode, err := requireField(node, "containerPort")
	if err != nil {
		return err
	}
	if portNode.Kind != yaml.ScalarNode {
		return Errorf(portNode, "containerPort must be int")
	}
	_, err = strconv.Atoi(portNode.Value)
	if err != nil {
		return Errorf(portNode, "containerPort must be int")
	}

	if protoNode := findMappingNode(node, "protocol"); protoNode != nil {
		if protoNode.Kind != yaml.ScalarNode {
			return Errorf(protoNode, "protocol must be string")
		}
		if protoNode.Value != "TCP" && protoNode.Value != "UDP" {
			return Errorf(protoNode, "protocol has unsupported value '%s'", protoNode.Value)
		}
	}

	return nil
}

func validateProbe(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return Errorf(node, "probe must be object")
	}

	httpGetNode, err := requireField(node, "httpGet")
	if err != nil {
		return err
	}
	return validateHTTPGet(httpGetNode)
}

func validateHTTPGet(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return Errorf(node, "httpGet must be object")
	}

	pathNode, err := requireField(node, "path")
	if err != nil {
		return err
	}
	if pathNode.Kind != yaml.ScalarNode {
		return Errorf(pathNode, "path must be string")
	}
	if !strings.HasPrefix(pathNode.Value, "/") {
		return Errorf(pathNode, "path has invalid format '%s'", pathNode.Value)
	}

	portNode, err := requireField(node, "port")
	if err != nil {
		return err
	}
	if portNode.Kind != yaml.ScalarNode {
		return Errorf(portNode, "port must be int")
	}
	_, err = strconv.Atoi(portNode.Value)
	if err != nil {
		return Errorf(portNode, "port must be int")
	}
	// НЕ проверяем диапазон порта в пробах

	return nil
}

func validateResources(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return Errorf(node, "resources must be object")
	}

	for _, section := range []string{"requests", "limits"} {
		if sectionNode := findMappingNode(node, section); sectionNode != nil {
			if sectionNode.Kind != yaml.MappingNode {
				return Errorf(sectionNode, "resources.%s must be object", section)
			}

			if cpuNode := findMappingNode(sectionNode, "cpu"); cpuNode != nil {
				if cpuNode.Kind != yaml.ScalarNode {
					return Errorf(cpuNode, "resources.%s.cpu must be int", section)
				}
				// Разрешаем любые значения (число или строка)
			}

			if memNode := findMappingNode(sectionNode, "memory"); memNode != nil {
				if memNode.Kind != yaml.ScalarNode {
					return Errorf(memNode, "resources.%s.memory must be string", section)
				}
				mem := memNode.Value
				if len(mem) < 3 {
					return Errorf(memNode, "resources.%s.memory has invalid format '%s'", section, mem)
				}
				suffix := mem[len(mem)-2:]
				if suffix != "Ki" && suffix != "Mi" && suffix != "Gi" {
					return Errorf(memNode, "resources.%s.memory has invalid format '%s'", section, mem)
				}
				numPart := mem[:len(mem)-2]
				if _, err := strconv.Atoi(numPart); err != nil {
					return Errorf(memNode, "resources.%s.memory has invalid format '%s'", section, mem)
				}
			}
		}
	}
	return nil
}