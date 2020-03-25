# Kube Descheduler Operator

Run the descheduler in your OpenShift cluster to move pods based on specific strategies.

## Deploy the operator

### Quick Development

1. Build and push the operator image to a registry:
2. Ensure the `image` spec in `deploy/05_deployment.yaml` refers to the operator image you pushed
3. Run `oc create -f deploy/.`

### OperatorHub install with custom index image

This process refers to building the operator in a way that it can be installed locally via the OperatorHub with a custom index image

1. build and push the image to a registry (e.g. https://quay.io):
   ```sh
   $ podman build -t quay.io/<username>/ose-cluster-kube-descheduler-operator-bundle:latest -f Dockerfile .
   $ podman push quay.io/<username>/ose-cluster-kube-descheduler-operator-bundle:latest
   ```

1. build and push image index for operator-registry (pull and build https://github.com/operator-framework/operator-registry/ to get the `opm` binary)
   ```sh
   $ ./bin/linux-amd64-opm index add --bundles quay.io/<username>/ose-cluster-kube-descheduler-operator-bundle:latest --tag quay.io/<username>/ose-cluster-kube-descheduler-operator-bundle-index:1.0.0
   $ podman push quay.io/<username>/ose-cluster-kube-descheduler-operator-bundle-index:1.0.0
   ```

   Don't forget to increase the number of open files, .e.g. `ulimit -n 100000` in case the current limit is insufficient.

1. create and apply catalogsource manifest:
   ```yaml
   apiVersion: operators.coreos.com/v1alpha1
   kind: CatalogSource
   metadata:
     name: cluster-kube-descheduler-operator
     namespace: openshift-marketplace
   spec:
     sourceType: grpc
     image: quay.io/<username>/ose-cluster-kube-descheduler-operator-bundle-index:1.0.0
   ```

1. create `cluster-kube-descheduler-operator` namespace:
   ```
   $ oc create ns cluster-kube-descheduler-operator
   ```

1. open the console Operators -> OperatorHub, search for `descheduler operator` and install the operator

## Sample CR

A sample CR definition looks like below (the operator expects `config` CR under `openshift-kube-descheduler-operator` namespace):

```yaml
apiVersion: operator.openshift.io/v1beta1
kind: KubeDescheduler
metadata:
  name: config
  namespace: openshift-kube-descheduler-operator
spec:
  deschedulingIntervalSeconds: 1800
  strategies:
    - name: "LowNodeUtilization"
      params:
       - name: "cputhreshold"
         value: "10"
       - name: "memorythreshold"
         value: "20"
       - name: "podsthreshold"
         value: "30"
       - name: "memorytargetthreshold"
         value: "40"
       - name: "cputargetthreshold"
         value: "50"
       - name: "podstargetthreshold"
         value: "60"
       - name: "nodes"
         value: "3"
    - name: "RemoveDuplicates"
```
The valid list of strategies are `RemoveDuplicates`, `LowNodeUtilization`, `RemovePodsViolatingInterPodAntiAffinity`, `RemovePodsViolatingNodeAffinity`, and `RemovePodsViolatingNodeTaints`. These strategies are documented in detail in the [descheduler README](https://github.com/kubernetes-sigs/descheduler/#policy-and-strategies).

Using the above strategies defined in CR we create a configmap in openshift-descheduler-operator namespace. As shown in the above example CR, the `LowNodeUtilization` strategy is the only one which accepts additional `params`, which map to the `thresholds` and `targetThresholds` parameters as defined in the [`LowNodeUtilization` section of the descheduler README](https://github.com/kubernetes-sigs/descheduler/#lownodeutilization). The `nodes` parameter corresponds to `numberOfNodes`, which activates this strategy only when the number of underutilized nodes is above the configured value (default `0`).

## How does the descheduler operator work?

Descheduler operator at a high level is responsible for watching the above CR
- Create a configmap that could be used by descheduler.
- Run descheduler as a deployment mounting the configmap as a policy file in the pod.

The configmap created from above sample CR definition looks like this:

```yaml
apiVersion: "kubedeschedulers.operator.openshift.io/v1beta1"
kind: "DeschedulerPolicy"
strategies:
  "RemoveDuplicates":
     enabled: true
  "LowNodeUtilization":
     enabled: true
     params:
       nodeResourceUtilizationThresholds:
         thresholds:
           "cpu" : 10
           "memory": 20
           "pods": 30
         targetThresholds:
           "cpu" : 40
           "memory": 50
           "pods": 60
         numberOfNodes: 3
```

The above configmap would be mounted as a volume in descheduler pod created. Whenever we change strategies, parameters or schedule in the CR, the descheduler operator is responsible for identifying those changes and regenerating the configmap. For more information on how descheduler works, please visit [descheduler](https://docs.openshift.com/container-platform/3.11/admin_guide/scheduling/descheduler.html)


## Parameters
The Descheduler operator exposes the following parameters in its CRD:

* `deschedulingIntervalSeconds` - this sets the number of seconds between descheduler runs
* `image` - specifies the Descheduler container image to deploy
* `flags` - this allows additional descheduler flags to be set, and they will be appended to the descheduler pod. Therefore, they must be in the same format as would be passed to the descheduler binary (eg, `"--dry-run"`)
