package cmd

import "fmt"

func Apply(manifest string) []string {
	return []string{
		"kubectl",
		"apply",
		"-f",
		manifest,
	}
}

func FluxInstall() []string {
	return []string{
		"flux",
		"install",
	}
}

func FluxReconcile(resource, name string) []string {
	return []string{
		"flux",
		"reconcile",
		resource,
		name,
	}
}

func GetNodeNames() []string {
	return []string{
		"kubectl",
		"get",
		"node",
		"-o",
		"name",
	}
}

func LabelNode(nodeName string) []string {
	return []string{
		"kubectl",
		"label",
		nodeName,
		"cncf-project=green-reviews",
		"cncf-project-sub=internal",
	}
}

func K3sManifests() []string {
	return []string{
		"/clusters/base/monitoring-namespace.yaml",
		"/clusters/base",
	}
}

func K3sPatches() [][]string {
	return [][]string{
		Patch("helmrelease",
			"kube-prometheus-stack",
			"flux-system",
			"/spec/values/prometheus-node-exporter",
			`{"hostRootFsMount": {"enabled": false}}`),
		Patch("helmrelease",
			"kepler",
			"flux-system",
			"/spec/values/canMount",
			`{"usrSrc": false}`),
	}
}

func Patch(resource, name, namespace, path, value string) []string {
	return []string{
		"kubectl",
		"patch",
		resource,
		name,
		"-n",
		namespace,
		"--type=json",
		"-p",
		fmt.Sprintf(`[{"op": "add", "path": "%s", "value": %s}]`, path, value),
	}
}

func WaitForNamespace(namespace string) []string {
	return []string{
		"kubectl",
		"wait",
		"pod",
		"--all",
		"--timeout",
		"300s",
		"--for",
		"condition=Ready",
		"--namespace",
		namespace,
	}
}
