# Tanzu Operator

## Kind - Sonobuoy

Creates a pod from which sonobuoy is run onto the cluster

To deploy this operator onto your cluster, run

```bash
make deploy IMG=rahulraju/test-tanzu-operator:0.1.0
```

After the CRDs and Controller is deployed, try creating a sample resource

```bash
kubectl apply -f ./config/samples/tanzu_v1alpha1_sonobuoy.yaml
```
