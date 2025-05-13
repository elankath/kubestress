package scenarios

import (
	"embed"
	"io"
	"io/fs"
	"log/slog"
	"path"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

//go:embed a/*
//go:embed *.yaml
var contentFS embed.FS

type ScenarioData struct {
	TemplatePods  []corev1.Pod
	TemplateNodes []corev1.Node
}

func LoadServiceAccount() (*corev1.ServiceAccount, error) {
	data, err := fs.ReadFile(contentFS, "sa-default.yaml")
	if err != nil {
		slog.Error("error when reading file", "name", "sa-default.yaml", "error", err)
		return nil, err
	}
	slog.Info("Loading spec", "name", "sa-default.yaml")
	stringReader := strings.NewReader(string(data))
	yamlDecoder := yaml.NewYAMLOrJSONDecoder(io.NopCloser(stringReader), 4096)
	var sa corev1.ServiceAccount
	err = yamlDecoder.Decode(&sa)
	if err != nil {
		return nil, err
	}
	return &sa, nil
}

func LoadScenario(name string) (s ScenarioData, err error) {
	// scheme := runtime.NewScheme()
	// corev1.AddToScheme(scheme)

	files, err := fs.ReadDir(contentFS, name)
	var data []byte
	if err != nil {
		return
	}
	slog.Info("Number of files in scenario", "scenario", name, "numfiles", len(files))
	for _, entry := range files {
		// if !entry.Type().IsRegular() {
		// 	continue
		// }
		data, err = fs.ReadFile(contentFS, path.Join(name, entry.Name()))
		if err != nil {
			slog.Error("error when reading file", "name", entry.Name(), "error", err)
			return
		}
		slog.Info("Loading spec", "name", entry.Name())
		stringReader := strings.NewReader(string(data))
		yamlDecoder := yaml.NewYAMLOrJSONDecoder(io.NopCloser(stringReader), 4096)

		if strings.HasPrefix(entry.Name(), "pod") {
			var pod corev1.Pod
			err = yamlDecoder.Decode(&pod)
			if err != nil {
				return
			}
			s.TemplatePods = append(s.TemplatePods, pod)
		} else if strings.HasPrefix(entry.Name(), "node") {
			var node corev1.Node
			err = yamlDecoder.Decode(&node)
			if err != nil {
				return
			}
			s.TemplateNodes = append(s.TemplateNodes, node)
		}

	}
	return
}
