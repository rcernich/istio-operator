package common

import (
	"github.com/maistra/istio-operator/pkg/apis/istio/v1alpha3"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/timeconv"
)

// RenderHelmChart renders the helm charts, returning a map of rendered templates.
func RenderHelmChart(chartPath string, icp *v1alpha3.IstioControlPlane, template string) (map[string]string, error) {
	rawVals, err := encode(icp.Spec)
	config := &chart.Config{Raw: string(rawVals), Values: map[string]*chart.Value{}}

	c, err := chartutil.Load(chartPath)
	if err != nil {
		panic(err)
	}

	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			// XXX: hard code or use icp.GetName()
			Name:      "istio",
			IsInstall: true,
			IsUpgrade: false,
			Time:      timeconv.Now(),
			Namespace: icp.GetNamespace(),
		},
		KubeVersion: kubeVersion
	}

	return renderutil.Render(c, config, renderOpts)
}

func encode(icp *v1alpha3.IstioControlPlane) ([]byte, error) {
	return []byte{}, nil
}
