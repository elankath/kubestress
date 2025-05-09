package api

const (
	ProgName = "kubestress"
)

type LoadConfig struct {
	KubeConfig   string
	ScenarioName string
	N            int
}
