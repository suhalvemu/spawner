package labels

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func merge(maps ...map[string]*string) map[string]*string {
	m := make(map[string]*string)

	for _, _m := range maps {
		for k, v := range _m {
			m[k] = v
		}
	}
	return m
}

func GetNodeLabel(nodeSpec *proto.NodeSpec) map[string]*string {

	instance := ""
	if nodeSpec.MachineType != "" {
		instance = nodeSpec.MachineType
		//+ is not allowed in tag value regex
		instance = strings.Replace(instance, "+", "-", 2)
	}
	if nodeSpec.Instance != "" {
		instance = nodeSpec.Instance
	}

	labels := map[string]*string{
		constants.NodeNameLabel:          &nodeSpec.Name,
		constants.InstanceLabel:          &instance,
		constants.NodeLabelSelectorLabel: &nodeSpec.Name,
		"type":                           aws.String("nodegroup")}

	return merge(DefaultTags(), labels, aws.StringMap(nodeSpec.Labels))
}

func ScopeTag() string {
	return fmt.Sprintf("nb-%s", config.Get().Env)
}

//DefaultTags labels/tags which is added to all spawner resources
func DefaultTags() map[string]*string {
	scope := ScopeTag()
	return map[string]*string{
		constants.Scope:        &scope,
		constants.CreatorLabel: &constants.SpawnerServiceLabel,
	}
}
