# Kubernetes storage (MySQL)

## PVC created by the StatefulSet

The MySQL `StatefulSet` uses a `volumeClaimTemplate` named `data`. For a StatefulSet named `mysql`, Kubernetes creates a PVC:

- `data-mysql-0` in namespace `asientos`

That PVC is the durable store mounted at `/var/lib/mysql`. It survives pod restarts and rescheduling as long as the PVC is retained.

## Optional static PersistentVolume

For clusters **without** a default dynamic provisioner, apply `k8s/mysql-pv.yaml` (hostPath example for lab use) and set on the StatefulSet volume claim template:

```yaml
storageClassName: manual
```

so the PVC binds to the PV with `storageClassName: manual`.

## Production note

Prefer a managed `StorageClass` (SSD, replication as offered by your cloud) instead of hostPath.
