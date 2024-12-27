package main

import (
	"context"
	"os"

	"dagger/dagger-test/internal/dagger"
	"dagger/dagger-test/pkg/pipeline"
)

const (
	clusterName = "green-reviews-test"
)

type DaggerTest struct{}

func (m *DaggerTest) BenchmarkPipeline(ctx context.Context,
	cncfProject,
	// +optional
	config,
	version,
	benchmarkJobURL,
	// +optional
	kubeconfig string,
	benchmarkJobDurationMins int) (*dagger.Container, error) {
	p, err := newPipeline(ctx, kubeconfig)
	if err != nil {
		return nil, err
	}
	if kubeconfig == "" {
		_, err = p.SetupCluster(ctx)
		if err != nil {
			return nil, err
		}
	}

	return p.Benchmark(ctx, cncfProject, config, version, benchmarkJobURL, benchmarkJobDurationMins)
}

func (m *DaggerTest) BenchmarkPipelineTest(ctx context.Context,
	// +optional
	// +default="falco"
	cncfProject,
	// +optional
	// +default="modern-ebpf"
	config,
	// +optional
	// +default="0.39.2"
	version,
	// +optional
	// +default="https://raw.githubusercontent.com/falcosecurity/cncf-green-review-testing/e93136094735c1a52cbbef3d7e362839f26f4944/benchmark-tests/falco-benchmark-tests.yaml"
	benchmarkJobURL,
	// +optional
	kubeconfig string,
	// +optional
	// +default=2
	benchmarkJobDurationMins int) (*dagger.Container, error) {
	p, err := newPipeline(ctx, kubeconfig)
	if err != nil {
		return nil, err
	}

	if kubeconfig == "" {
		_, err = p.SetupCluster(ctx)
		if err != nil {
			return nil, err
		}
	}

	return p.Benchmark(ctx,
		cncfProject,
		config,
		version,
		benchmarkJobURL,
		benchmarkJobDurationMins)
}

func (m *DaggerTest) SetupCluster(ctx context.Context,
	// +optional
	kubeconfig string) (*dagger.Container, error) {
	p, err := newPipeline(ctx, kubeconfig)
	if err != nil {
		return nil, err
	}

	return p.SetupCluster(ctx)
}

func newPipeline(ctx context.Context, kubeconfig string) (*pipeline.Pipeline, error) {
	var configFile *dagger.File
	var err error

	dir := dag.CurrentModule().Source()
	container := build(dir)

	if kubeconfig == "" {
		configFile, err = startK3sCluster(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		configFile, err = getKubeconfig(kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	return pipeline.New(container, dir, configFile)
}

func build(src *dagger.Directory) *dagger.Container {
	return dag.Container().
		WithDirectory("/src", src).
		WithWorkdir("/src").
		Directory("/src").
		DockerBuild()
}

func getKubeconfig(configFilePath string) (*dagger.File, error) {
	contents, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	filePath := "/.kube/config"
	dir := dag.Directory().WithNewFile(filePath, string(contents))
	return dir.File(filePath), nil
}

func startK3sCluster(ctx context.Context) (*dagger.File, error) {
	k3s := dag.K3S(clusterName)
	kServer := k3s.Server()

	kServer, err := kServer.Start(ctx)
	if err != nil {
		return nil, err
	}
	return k3s.Config(), nil
}
