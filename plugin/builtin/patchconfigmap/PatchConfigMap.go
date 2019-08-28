//go:generate go run sigs.k8s.io/kustomize/v3/cmd/pluginator
package main

import (
	"errors"
	// "regexp"

	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resid"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/transformers"
)

type plugin struct{}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

const preStop = "- name: istio-proxy\\n" +
	"    lifecycle:\\n" +
	"      preStop:\\n" +
	"        exec:\\n" +
	"          command: [\"sh\", \"-c\", 'sleep 20; while [ $(netstat -plunt | grep tcp | grep -v envoy | wc -l | xargs) -ne 0 ]; do sleep 1; done']\\n"

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) error {
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	cmGVK := gvk.Gvk{
		Kind:    "ConfigMap",
		Version: "v1",
	}
	// Edit sidecar injector template
	sidecarInjectorCMId := resid.NewResIdWithNamespace(
		cmGVK, "istio-sidecar-injector", "istio-system")
	err := errors.New("cannot find istio-sidecar-injector config map")
	for _, r := range m.Resources() {
		if !sidecarInjectorCMId.Equals(r.OrgId()) {
			continue
		}
		err = transformers.MutateField(
			r.Map(), []string{"data", "config"},
			/* CreateIfNotPresent = */ false,
			p.editSidecarInjectorTemplate)
	}
	if err != nil {
		return err
	}

	// add other config map patch
	return nil
}

func (p *plugin) editSidecarInjectorTemplate(in interface{}) (interface{}, error) {
	return in, nil
	// if config, ok := in.(string); ok {
	// 	// re := regexp.MustCompile(`- name: istio-proxy\\n`)
	// 	if m := re.Find([]byte(config)); len(m) == 0 {
	// 		return nil, fmt.Errorf("cannot find `- name: istio-proxy\\n` in istio-sidecar-injector template %#v", in)
	// 	}
	// 	return re.ReplaceAllString(config, preStop), nil
	// }
	// return nil, fmt.Errorf("%#v is expected to be string", in)
}
