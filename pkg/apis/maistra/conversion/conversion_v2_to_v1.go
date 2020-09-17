package conversion

import (
	"fmt"

	conversion "k8s.io/apimachinery/pkg/conversion"

	"github.com/maistra/istio-operator/pkg/apis/maistra/status"
	v1 "github.com/maistra/istio-operator/pkg/apis/maistra/v1"
	v2 "github.com/maistra/istio-operator/pkg/apis/maistra/v2"
)

func v2ToV1Hacks(values map[string]interface{}, out *v1.ControlPlaneSpec) error {
	// adjustments for 3scale
	// Need to move 3scale out of Istio values into ThreeScale field
	if rawThreeScaleValues, ok := values["3scale"]; ok && rawThreeScaleValues != nil {
		if threeScaleValues, ok := rawThreeScaleValues.(map[string]interface{}); ok {
			out.ThreeScale = v1.NewHelmValues(threeScaleValues)
		} else {
			return fmt.Errorf("could not convert 3scale values to map[string]interface{}")
		}
	}
	delete(values, "3scale")

	hv := v1.NewHelmValues(values)
	rawJaegerValues, ok, err := hv.GetFieldNoCopy("tracing.jaeger")
	if ok {
		jaegerValues, ok := rawJaegerValues.(map[string]interface{})
		if !ok {
			return fmt.Errorf("could not cast tracing.jaeger value to map[string]interface{}: %T", rawJaegerValues)
		}
		// move tracing.jaeger.annotations to tracing.jaeger.podAnnotations
		if jaegerAnnotations, ok, err := hv.GetFieldNoCopy("tracing.jaeger.podAnnotations"); ok {
			if err := hv.SetField("tracing.jaeger.annotations", jaegerAnnotations); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		// normalize jaeger images
		if agentImage, ok, err := hv.GetString("tracing.jaeger.agent.image"); ok {
			if err := hv.SetField("tracing.jaeger.agentImage", agentImage); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if allInOneImage, ok, err := hv.GetString("tracing.jaeger.allInOne.image"); ok {
			if err := hv.SetField("tracing.jaeger.allInOneImage", allInOneImage); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if collectorImage, ok, err := hv.GetString("tracing.jaeger.collector.image"); ok {
			if err := hv.SetField("tracing.jaeger.collectorImage", collectorImage); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if queryImage, ok, err := hv.GetString("tracing.jaeger.query.image"); ok {
			if err := hv.SetField("tracing.jaeger.queryImage", queryImage); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		delete(jaegerValues, "podAnnotations")
		delete(jaegerValues, "agent")
		delete(jaegerValues, "allInOne")
		delete(jaegerValues, "collector")
		delete(jaegerValues, "query")
	} else if err != nil {
		return err
	}

	return nil
}

// Convert_v2_ControlPlaneSpec_To_v1_ControlPlaneSpec converts a v2 ControlPlaneSpec to an equivalent values.yaml.
// XXX: this requires the following additional details:
//      * namespace - the target namespace for the resource
func Convert_v2_ControlPlaneSpec_To_v1_ControlPlaneSpec(in *v2.ControlPlaneSpec, out *v1.ControlPlaneSpec, s conversion.Scope) error {
	if err := autoConvert_v2_ControlPlaneSpec_To_v1_ControlPlaneSpec(in, out, s); err != nil {
		return err
	}

	if len(in.Profiles) == 1 {
		out.Template = in.Profiles[0]
	}

	// Make a copy so we can modify fields as needed
	in = in.DeepCopy()

	// Initialize output
	values := make(map[string]interface{})

	// Cluster settings
	// cluster must come first as it may modify other settings on the input (e.g. meshExpansionPorts)
	if err := populateClusterValues(in, values); err != nil {
		return err
	}

	// General
	if err := populateGeneralValues(in.General, values); err != nil {
		return err
	}

	// Policy
	if err := populatePolicyValues(in, values); err != nil {
		return err
	}

	// Proxy
	if err := populateProxyValues(in, values); err != nil {
		return err
	}

	// Security
	if err := populateSecurityValues(in, values); err != nil {
		return err
	}

	// Telemetry
	if err := populateTelemetryValues(in, values); err != nil {
		return err
	}

	// Gateways
	if err := populateGatewaysValues(in, values); err != nil {
		return err
	}

	// Addons
	if err := populateAddonsValues(in, values); err != nil {
		return err
	}

	// Runtime - must run last as this will add values to existing child maps
	if err := populateControlPlaneRuntimeValues(in.Runtime, values); err != nil {
		return err
	}

	if err := v2ToV1Hacks(values, out); err != nil {
		return err
	}

	out.Istio = v1.NewHelmValues(values)

	return nil
}

func Convert_v2_ControlPlaneStatus_To_v1_ControlPlaneStatus(in *v2.ControlPlaneStatus, out *v1.ControlPlaneStatus, s conversion.Scope) error {
	// ReconciledVersion requires manual conversion: does not exist in peer-type
	out.ReconciledVersion = status.ComposeReconciledVersion(in.OperatorVersion, in.ObservedGeneration)
	// LastAppliedConfiguration requires manual conversion: does not exist in peer-type
	in.AppliedValues.DeepCopyInto(&out.LastAppliedConfiguration)
	return nil
}
