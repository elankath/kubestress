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

//go:embed a/pod-p.yaml
//go:embed a/node-p.yaml
var content embed.FS

type ScenarioData struct {
	PodSpecs  []corev1.Pod
	NodeSpecs []corev1.Node
}

func LoadScenario(name string) (s ScenarioData, err error) {
	// scheme := runtime.NewScheme()
	// corev1.AddToScheme(scheme)

	files, err := fs.ReadDir(content, name)
	var data []byte
	if err != nil {
		return
	}
	slog.Info("Number of files in scenario", "scenario", name, "numfiles", len(files))
	for _, entry := range files {
		// if !entry.Type().IsRegular() {
		// 	continue
		// }
		data, err = fs.ReadFile(content, path.Join(name, entry.Name()))
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
			s.PodSpecs = append(s.PodSpecs, pod)
		} else if strings.HasPrefix(entry.Name(), "node") {
			var node corev1.Node
			err = yamlDecoder.Decode(&node)
			if err != nil {
				return
			}
			s.NodeSpecs = append(s.NodeSpecs, node)
		}

	}
	return
}
