package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/elankath/kubestress/api"
	"github.com/elankath/kubestress/cli"
	"github.com/elankath/kubestress/core"
	flag "github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("Expected one of '%s load|cleanup sub-commands", api.ProgName))
		os.Exit(cli.ExitBasicInvocation)
	}

	var exitCode int
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	info, ok := debug.ReadBuildInfo()
	if ok {
		fmt.Printf("App version: %s\n", info.Main.Version)
	} else {
		fmt.Println("Warn: binary build info not embedded")
	}

	command := os.Args[1]
	switch command {
	case "load":
		exitCode, err = ExecLoad(ctx)
	case "cleanup":
		exitCode, err = ExecCleanup(ctx)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "%s: error: Unknown subcommand %q\n", api.ProgName, command)
		os.Exit(cli.ExitBasicInvocation)
	}
	if exitCode > 0 {
		slog.Error("error", command, err)
		os.Exit(exitCode)
	}
	slog.Info("DONE.", "command", command)
}

func ExecLoad(ctx context.Context) (exitCode int, err error) {
	var loadConfig api.LoadConfig
	loadFlags := flag.NewFlagSet("load", flag.ContinueOnError)
	loadFlags.StringVarP(&loadConfig.KubeConfig, clientcmd.RecommendedConfigPathFlag, "k", os.Getenv(clientcmd.RecommendedConfigPathEnvVar), "kubeconfig path of target cluster - defaults to KUBECONFIG env-var")
	loadFlags.StringVarP(&loadConfig.ScenarioName, "scenario", "s", os.Getenv("SCENARIO"), "name of load scenario - defaults to SCENARIO env-var")
	loadFlags.IntVarP(&loadConfig.N, "number", "n", 1, "Number of repeats of scenario - defaults to 1")
	standardUsage := loadFlags.PrintDefaults
	loadFlags.Usage = func() {
		_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("Usage: %s load <flags> <args>", api.ProgName))
		_, _ = fmt.Fprintln(os.Stderr, "<flags>")
		standardUsage()
		_, _ = fmt.Fprintln(os.Stderr, "<args>: <To be implemented>")
	}

	err = loadFlags.Parse(os.Args[2:])
	if err != nil {
		exitCode = cli.ExitBasicInvocation
		return
	}

	if loadConfig.ScenarioName == "" {
		err = fmt.Errorf("scenario name is mandatory")
		exitCode = cli.ExitMissingNecessaryArgument
		return
	}

	if loadConfig.KubeConfig == "" {
		err = fmt.Errorf("kubeconfig is mandatory")
		exitCode = cli.ExitMissingNecessaryArgument
		return
	}
	if loadConfig.N == 0 {
		err = fmt.Errorf("should have atleast one replica count")
		exitCode = cli.ExitInvalidArg
		return
	}
	loader, err := core.NewLoader(loadConfig)

	if err != nil {
		exitCode = cli.ExitCreateServices
		return
	}

	slog.Info("Loader config is ", "loadConfig", loadConfig)

	err = loader.Execute(ctx)
	if err != nil {
		exitCode = cli.ExitExecuteLoader
		return
	}
	return
}

func ExecCleanup(ctx context.Context) (exitCode int, err error) {
	return
}
