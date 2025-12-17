// validator/rules.go
package validator

import (
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var snakeCaseRegex = regexp.MustCompile(`^[a-z0-9_]+$`)

func validateTopLevel(root *yaml.Node) error {
	if root.Kind != yaml.MappingNode {
		return Errorf(root, "root must be a mapping")
	}

	apiVersion, err := requireField(root, "apiVersion")
	if err != nil {
		return err
	}
	if apiVersion.Kind != yaml.ScalarNode || apiVersion.Value != "v1" {
		return Errorf(apiVersion, "apiVersion has unsupported value '"+apiVersion.Value+"'")
	}

	kind, err := requireField(root, "kind")
	if err != nil {
		return err
	}
	if kind.Kind != yaml.ScalarNode || kind.Value != "Pod" {
		return Errorf(kind, "kind has unsupported value '"+kind.Value+"'")
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
	// Пустая строка или null — разрешено
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

	for _, cont := range containersNode.Content {
		if err := validateContainer(cont); err != nil {
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
	val := nameNode.Value
	if val != "linux" && val != "windows" {
		return Errorf(nameNode, "spec.os.name has unsupported value '"+val+"'")
	}
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
	if name := nameNode.Value; !snakeCaseRegex.MatchString(name) || name == "" {
		return Errorf(nameNode, "containers.name has invalid format '"+name+"'")
	}

	imageNode, err := requireField(node, "image")
	if err != nil {
		return err
	}
	if imageNode.Kind != yaml.ScalarNode {
		return Errorf(imageNode, "containers.image must be string")
	}
	image := imageNode.Value
	if !strings.HasPrefix(image, "registry.bigbrother.io/") || !strings.Contains(image, ":") {
		return Errorf(imageNode, "containers.image has invalid format '"+image+"'")
	}

	if ports := findMappingNode(node, "ports"); ports != nil {
		if ports.Kind != yaml.SequenceNode {
			return Errorf(ports, "containers.ports must be array")
		}
		for _, p := range ports.Content {
			if err := validatePort(p); err != nil {
				return err
			}
		}
	}

	for _, probe := range []string{"readinessProbe", "livenessProbe"} {
		if p := findMappingNode(node, probe); p != nil {
			if err := validateProbe(p); err != nil {
				return err
			}
		}
	}

	resNode, err := requireField(node, "resources")
	if err != nil {
		return err
	}
	return validateResources(resNode)
}

func validatePort(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return Errorf(node, "port item must be object")
	}

	portNode, err := requireField(node, "containerPort")
	if err != nil {
		return err
	}
	if portNode.Kind != yaml.ScalarNode {
		return Errorf(portNode, "containerPort must be int")
	}
	p, err2 := strconv.Atoi(portNode.Value)
	if err2 != nil || p <= 0 || p >= 65536 {
		return Errorf(portNode, "containerPort value out of range")
	}

	if proto := findMappingNode(node, "protocol"); proto != nil {
		if proto.Kind != yaml.ScalarNode {
			return Errorf(proto, "protocol must be string")
		}
		if proto.Value != "TCP" && proto.Value != "UDP" {
			return Errorf(proto, "protocol has unsupported value '"+proto.Value+"'")
		}
	}
	return nil
}

func validateProbe(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return Errorf(node, "probe must be object")
	}

	getNode, err := requireField(node, "httpGet")
	if err != nil {
		return err
	}
	return validateHTTPGet(getNode)
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
		return Errorf(pathNode, "path has invalid format '"+pathNode.Value+"'")
	}

	portNode, err := requireField(node, "port")
	if err != nil {
		return err
	}
	if portNode.Kind != yaml.ScalarNode {
		return Errorf(portNode, "port must be int")
	}
	p, err2 := strconv.Atoi(portNode.Value)
	if err2 != nil || p <= 0 || p >= 65536 {
		return Errorf(portNode, "port value out of range")
	}
	return nil
}

func validateResources(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return Errorf(node, "resources must be object")
	}

	for _, sec := range []string{"requests", "limits"} {
		if secNode := findMappingNode(node, sec); secNode != nil {
			if secNode.Kind != yaml.MappingNode {
				return Errorf(secNode, "resources."+sec+" must be object")
			}

			if cpu := findMappingNode(secNode, "cpu"); cpu != nil {
				if cpu.Kind != yaml.ScalarNode {
					return Errorf(cpu, "resources."+sec+".cpu must be int")
				}
				if _, err := strconv.Atoi(cpu.Value); err != nil {
					return Errorf(cpu, "resources."+sec+".cpu must be int")
				}
			}

			if mem := findMappingNode(secNode, "memory"); mem != nil {
				if mem.Kind != yaml.ScalarNode {
					return Errorf(mem, "resources."+sec+".memory must be string")
				}
				m := mem.Value
				if len(m) < 3 {
					return Errorf(mem, "resources."+sec+".memory has invalid format '"+m+"'")
				}
				suf := m[len(m)-2:]
				if suf != "Ki" && suf != "Mi" && suf != "Gi" {
					return Errorf(mem, "resources."+sec+".memory has invalid format '"+m+"'")
				}
				if _, err := strconv.Atoi(m[:len(m)-2]); err != nil {
					return Errorf(mem, "resources."+sec+".memory has invalid format '"+m+"'")
				}
			}
		}
	}
	return nil
}