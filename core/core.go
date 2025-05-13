package core

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/elankath/kubestress/api"
	"github.com/elankath/kubestress/core/scenarios"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Loader struct {
	cfg          api.LoadConfig
	client       *kubernetes.Clientset
	scenarioData scenarios.ScenarioData
	defaultSA    *corev1.ServiceAccount
}

func NewLoader(cfg api.LoadConfig) (loader *Loader, err error) {
	clientCfg, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
	if err != nil {
		err = fmt.Errorf("%w: %w", api.ErrCreateLoader, err)
		return
	}
	loader = &Loader{
		cfg: cfg,
	}
	loader.scenarioData, err = scenarios.LoadScenario(loader.cfg.ScenarioName)
	loader.client, err = kubernetes.NewForConfig(clientCfg)
	if err != nil {
		err = fmt.Errorf("%w: %w", api.ErrCreateLoader, err)
		return
	}
	loader.defaultSA, err = scenarios.LoadServiceAccount()
	if err != nil {
		err = fmt.Errorf("%w: %w", api.ErrCreateLoader, err)
		return
	}
	return
}

func (l *Loader) createDefaultSA(ctx context.Context) error {
	_, err := l.client.CoreV1().ServiceAccounts("default").Create(ctx, l.defaultSA, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("can not create default service account: %w", err)
	}
	return nil
}

func (l *Loader) Execute(ctx context.Context) (err error) {
	err = l.createDefaultSA(ctx)
	if err != nil {
		return
	}
	for i := range l.cfg.N {
		var node *corev1.Node
		var pod *corev1.Pod
		for _, noSpec := range l.scenarioData.NodeSpecs {
			node, err = l.client.CoreV1().Nodes().Create(ctx, &noSpec, metav1.CreateOptions{})
			if err != nil {
				err = fmt.Errorf("cannot create node %q: %w", node.GetGenerateName(), err)
				return
			}
			err = adjustNode(l.client, node.Name, node.Status)
			if err != nil {
				return
			}

			slog.Info("Created node", "scenarioName", l.cfg.ScenarioName, "nodeName", node.Name, "count", i)
		}
		for _, poSpec := range l.scenarioData.PodSpecs {
			// slog.Info("Creating pod", "generateName", poSpec.GetGenerateName())
			if poSpec.Namespace == "" {
				poSpec.Namespace = "default"
			}
			pod, err = l.client.CoreV1().Pods(poSpec.Namespace).Create(ctx, &poSpec, metav1.CreateOptions{})
			if err != nil {
				err = fmt.Errorf("cannot create pod %q: %w", pod.GetGenerateName(), err)
				return
			}
			slog.Info("Created pod", "scenarioName", l.cfg.ScenarioName, "podName", pod.Name, "count", i)
		}
	}
	return nil
}

// BuildReadyConditions sets up mock NodeConditions
func BuildReadyConditions() []corev1.NodeCondition {
	lastTransition := time.Now().Add(-time.Minute)
	return []corev1.NodeCondition{
		{
			Type:               corev1.NodeReady,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Time{Time: lastTransition},
		},
		{
			Type:               corev1.NodeNetworkUnavailable,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Time{Time: lastTransition},
		},
		{
			Type:               corev1.NodeDiskPressure,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Time{Time: lastTransition},
		},
		{
			Type:               corev1.NodeMemoryPressure,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Time{Time: lastTransition},
		},
	}
}

func adjustNode(clientSet *kubernetes.Clientset, nodeName string, nodeStatus corev1.NodeStatus) error {

	nd, err := clientSet.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("adjustNode cannot get node with name %q: %w", nd.Name, err)
	}
	nd.Status.Conditions = BuildReadyConditions()
	nd.Spec.Taints = lo.Filter(nd.Spec.Taints, func(item corev1.Taint, index int) bool {
		return item.Key != "node.kubernetes.io/not-ready"
	})
	nd, err = clientSet.CoreV1().Nodes().Update(context.Background(), nd, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("adjustNode cannot update node with name %q: %w", nd.Name, err)
	}
	nd.Status = nodeStatus
	nd.Status.Phase = corev1.NodeRunning
	nd, err = clientSet.CoreV1().Nodes().UpdateStatus(context.Background(), nd, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("adjustNode cannot update the status of node with name %q: %w", nd.Name, err)
	}
	return nil
}
