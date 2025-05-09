package core

import (
	"context"
	"fmt"
	"github.com/elankath/kubestress/api"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Loader struct {
	cfg    api.LoadConfig
	client *kubernetes.Clientset
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
	loader.client, err = kubernetes.NewForConfig(clientCfg)
	if err != nil {
		err = fmt.Errorf("%w: %w", api.ErrCreateLoader, err)
		return
	}
	return
}

func (l *Loader) Execute(ctx context.Context) (err error) {
	return nil
}
