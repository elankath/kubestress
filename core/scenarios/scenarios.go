package scenarios

import "embed"

//go:embed a/pod-p.yaml
//go:embed a/node-p.yaml
var content embed.FS

type Scenario struct {
	PodSpecPaths  []string
	NodeSpecPaths []string
}

func LoadScenario(name string) (*Scenario, error) {
	return nil, nil
}
