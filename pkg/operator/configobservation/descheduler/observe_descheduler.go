package descheduler

import (
	"github.com/openshift/cluster-kube-descheduler-operator/pkg/operator/configobservation"
	"github.com/openshift/cluster-kube-descheduler-operator/pkg/operator/operatorclient"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

func ObserveDeschedulerConfig(genericListers configobserver.Listers, recorder events.Recorder, existingConfig map[string]interface{}) (map[string]interface{}, []error) {
	listers := genericListers.(configobservation.Listers)
	errs := []error{}
	prevObservedConfig := map[string]interface{}{}

	observedConfig := map[string]interface{}{}
	deschedulerConfig, err := listers.DeschedulerLister.KubeDeschedulers(operatorclient.OperatorNamespace).Get("cluster")
	if errors.IsNotFound(err) {
		klog.Warningf("deschedulers/cluster: not found")
		return observedConfig, errs
	}
	if err != nil {
		errs = append(errs, err)
		return prevObservedConfig, errs
	}
	configMapName := deschedulerConfig.Spec.Policy.Name
	configMap, err := listers.ConfigmapLister.ConfigMaps(operatorclient.OperatorNamespace).Get(configMapName)
	if errors.IsNotFound(err) {
		klog.Warningf("descheduler policy configmap '%s' not found", configMapName)
		return observedConfig, errs
	}
	if err != nil {
		errs = append(errs, err)
		return prevObservedConfig, errs
	}

	unstructured.NestedMap(configMap.Data["policy.yaml"], "strategies")
	err := unstructured.SetNestedMap(observedConfig, configMap.Data["policy.yaml"])
	return observedConfig, errs
}
