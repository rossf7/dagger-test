package pipeline

import (
	"context"
	"dagger/dagger-test/internal/dagger"
	"dagger/dagger-test/pkg/cmd"
	"path"
)

type Pipeline struct {
	container  *dagger.Container
	dir        *dagger.Directory
	kubeConfig *dagger.File
}

func New(container *dagger.Container, dir *dagger.Directory, kubeConfig *dagger.File) Pipeline {
	return Pipeline{
		container:  container,
		dir:        dir,
		kubeConfig: kubeConfig,
	}
}

func (p *Pipeline) SetupCluster(ctx context.Context) (*dagger.Container, error) {
	var err error
	/*
		stdout, err := m.runCmd(ctx, k3s.Config(), cmd.GetNodeNames())
		if err != nil {
			return nil, err
		}
		// return m.terminal(ctx, k3s.Config())

		node := strings.Split(stdout, "\n")[0]
		_, err = m.runCmd(ctx, k3s.Config(), cmd.LabelNode(node))
		if err != nil {
			return nil, err
		}

		return m.terminal(ctx, k3s.Config())
	*/

	_, err = p.runCmd(ctx, cmd.FluxInstall())
	if err != nil {
		return nil, err
	}

	for _, manifest := range k3sManifests() {
		_, err = p.runCmdWithFile(ctx, manifest, cmd.Apply(manifest))
		if err != nil {
			return nil, err
		}
	}

	for _, patch := range k3sPatches() {
		_, err = p.runCmd(ctx, patch)
		if err != nil {
			return nil, err
		}
	}

	_, err = p.runCmd(ctx, cmd.FluxReconcile("helmrelease", "kepler"))
	if err != nil {
		return nil, err
	}

	_, err = p.runCmd(ctx, cmd.WaitForNamespace("monitoring"))
	if err != nil {
		return nil, err
	}

	// return p.Terminal(ctx)

	return nil, nil
}

func (p *Pipeline) runCmd(ctx context.Context, args []string) (string, error) {
	stdout, err := p.container.
		WithEnvVariable("KUBECONFIG", "/.kube/config").
		WithFile("/.kube/config", p.kubeConfig).
		WithExec(args).
		Stdout(ctx)
	if err != nil {
		return "", err
	}
	return stdout, nil
}

func (p *Pipeline) runCmdWithFile(ctx context.Context, manifestPath string, args []string) (string, error) {
	dirPath := path.Dir(manifestPath)
	dir := p.dir.Directory(dirPath)
	stdout, err := p.container.
		WithEnvVariable("KUBECONFIG", "/.kube/config").
		WithFile("/.kube/config", p.kubeConfig).
		WithDirectory(dirPath, dir).
		WithExec(args).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	return stdout, nil
}

func (p *Pipeline) Terminal(ctx context.Context) (*dagger.Container, error) {
	return p.container.WithEnvVariable("KUBECONFIG", "/.kube/config").
		WithFile("/.kube/config", p.kubeConfig).Terminal(), nil
}

func k3sManifests() []string {
	return []string{
		"/clusters/base/monitoring-namespace.yaml",
		"/clusters/base",
	}
}

func k3sPatches() [][]string {
	return [][]string{
		cmd.Patch("helmrelease",
			"kube-prometheus-stack",
			"flux-system",
			"/spec/values/prometheus-node-exporter",
			`{"hostRootFsMount": {"enabled": false}}`),
		cmd.Patch("helmrelease",
			"kepler",
			"flux-system",
			"/spec/values/canMount",
			`{"usrSrc": false}`),
	}
}
