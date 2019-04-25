# Persistent Volumes and Persistent Volume Claims in Kubernetes

## Overview (pv available in the cluster, pvc exists in a namespace)

[image]

PVs are created by cluster admins and they are consumed by PVCs created by developers.
A PV is like a mount specification to a storage. Therefore, you can create multiple PVs with different mount
specification for the same storage.
A PV is public resource in a cluster, which means it is accessible to all the namespace. This also means
a PV's name needs to be unique in the whole cluster.
A PVC is an object within a namespace. So its name must be unique in the namespace.
A PV can only be exclusively bound to a PVC. This one-to-one mapping lasts until the PVC is deleted.

## Provisioning

There are two ways to provision PVs: statically or dynamically.

### "Static" PV

A static PV means it is manually created by a cluster administrator with details of real storage.
"Static" here means it must exist before being consumed by a PVC.

Here is an example of static PVs:

PV spec:
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-pv
spec:
  nfs:
    # FIXME: use the right IP
    server: 12.34.56.78
    path: "/data"
    readOnly: false
  mountOptions:
    - vers=4.0
    - rsize=32768
    - wsize=32768
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
```

PVC spec:
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs-pvc
spec:
  resources:
    requests:
      storage: 10Gi
  accessModes:
    - ReadWriteMany
  selector:
    matchLabels:
      pv-name: nfs-pv
```

Pod spec:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nfs-pod
  labels:
    app: nfs-pods
spec:
  volumes:
  - name: data-dir
    persistentVolumeClaim:
      claimName: nfs-pvc

  containers:
  - name: nginx
    image: nginx
    volumeMounts:
      - name: data-dir
        mountPath: "/usr/share/nginx/html"
        readOnly: false
```

In this example, Pod `nfs-pod` utilizes PVC `nfs-pvc` to create a volume called `data-dir` and then mounts the volume
to the directory `/usr/share/nginx/html` in the container `nginx`. The PVC `ngx-pvc` finds and binds to the PV `nfs-pv` via
[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#labelselector-v1-meta). After
creation of these Kubernetes objects, ssh into container `nginx` and run `cat /proc/mount` then you will find info
of PV `nfs-pv`, like this: `12.34.56.78:/data /usr/share/nginx/html nfs4 vers=4.0,rsize=32768,wsize=32768,...,addr=12.34.56.78 0 0`.
This means the directory `/usr/share/nginx/html` is mapped to `12.34.56.78:/data`.

#### PV Types and Mount Options
Kubernetes currently supports a lot of PV types, for example NFS, CephFS, Glusterfs and GCEPersistentDisk. You can check
[PersistentVolumeSpec] for details of PV types.

In this example, the persistent volume `nfs-pv` is created using NFS PV type, with server `12.34.56.78` and `path`. Moreover,
the PV also specify some mount options for connecting to the NFS server. Mount options are only supported by some PV types,
check [this doc](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#mount-options) for more details.


#### The capacity of A PV
The capacity of a static PV is not hard limit of real storage. Instead, the capacity is fully controlled by the real storage.
Therefore, suppose the NFS server in the example has 200Gi storage space, PV `nfs-pv` is able to use up all of the NFS server's
space even although it only has `capacity.storage == 10Gi`. Additionally, The capacity setting of a static PV normally is just for
matching up storage request in the corresponding PVC.

#### Access Mode
There are three access modes for PVs and PVCs:
- `ReadWriteOnce`: a PV can be mounted as read-write by a single node if it has `ReadWriteOnce` in its `accessModes` spec.
This means: 1. the PV can perform read and write to real storage behind it. 2. The PV can only be mounted
on a single node, which means any Pod that wants to use this PV must be scheduled to this node as well.
- `ReadOnlyMany`: a PV can be mounted as read-only by many nodes if it has `ReadOnlyMany` in its `accessModes` spec.
Unlike `ReadWriteOnce`, `ReadOnlyMany` allows the PV to be mounted on many nodes although it can only perform read to
real storage.
- `ReadWriteMany`: a PV can be mounted as read-write by many nodes if it has `ReadWriteMany` in its `accessModes` spec.

Different PV types have different supports for these three accessModes. You can check
[this doc](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes) for more details.

You may notice that the PV's `accessModes` field is an array, which means it can has multiple accessModes. **Nevertheless,
a PV can only be mounted using one access mode at a time, even if it has many in its `accessModes` field. Therefore,
instead of supporting multiple access modes in a PV, it is recommended to support merely one access mode in one PV and create
separate PVs with different access modes for different purpose.**

You may also notice that there are other attributes which can also affect access modes. Here is simplified summary:
- `readOnly` attribute of a PV type is storage side setting. It is used to control whether real storage is read-only or not.
- `AccessModes` of a PV is PV side setting and it is used to control access mode of the PV. It builds the bridge between
"clients" and real storage.
- `AccessModes` of a PVC has to match up the PV it wants to bind. PV and PVC build a bridge between the "client" and
real storage: PV connects to real storage while PVC connects to the "client".
- `readOnly` attribute of [VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#nfsvolumesource-v1-core)
is "client" side setting. It is used to control whether the mounted directory is read-only or not.

#### Reclaim Policy
The field `persistentVolumeReclaimPolicy` specifies reclaim policy for the PV, which can be either `Delete` (default value) or `Retain`.
You need to set it `Retain` and back up data in the PV yourself in case of data lost.

#### Binding
The example above uses [LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#labelselector-v1-meta)
`matchLabels.pv-name == pv-name` to bind PV `nfs-pv` and PVC `nfs-pvc` together. You do not need to use LabelSelector
to establish the bind between a PV and your PVC if you want more flexible way of binding. For example, without LabelSelector,
a PVC which requires `storage == 10Gi` and `accessModes == [ReadWriteOnce]` can be bound to a PV with `storage >= 10Gi`
and `accessModes includes ReadWriteOnce like [ReadWriteOnce, ReadWriteMany]`.

### "Dynamic" Persistent Volumes

Dynamic Persistent Volumes are the volumes which are dynamically created by the cluster with specification of a
user's PVC. This provisioning is based on `StorageClasses`: the PVC must specify an existing [storage class] in
order to create a dynamic PV.

#### Storage Classes.

A `StorageClass` is a Kubernetes object used to describe a class of storage. It uses fields like `parameters`, `provisioner`
and `reclaimPolicy` to describe details about the storage class it represents. Let's take the default `StorageClass` `standard`
in the GKE K8s Cluster as en example, here is its spec:

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: standard
parameters:
  type: pd-standard
provisioner: kubernetes.io/gce-pd
reclaimPolicy: Delete
volumeBindingMode: Immediate
```

Explanation of these fields:
- The field `metadata.name` is the name of the `StorageClass`. It has to be unique in the whole cluster.
- The field `parameters` specifies parameters for real storage. For example `parameters.type == pd-standard` means
the `StorageClass` uses [GCEPersistentDisk] as storage media. You can check [this doc](https://kubernetes.io/docs/concepts/storage/storage-classes/#parameters)
for more details.
-  The field `provisioner` specifies which volume plugin is used for provisioning dynamic PVs for this `StorageClass`.
You can check [this list](https://kubernetes.io/docs/concepts/storage/storage-classes/#provisioner) for each
provisioner's specification.
- Like `persistentVolumeReclaimPolicy`, the field `reclaimPolicy` sets up reclaim policy for the `StorageClass`,
which can be either `Delete` (default value) or `Retain`.
- The field `volumeBindingMode` controls when to do dynamic provisioning for the PV and volume binding.
`volumeBindingMode == Immediate` means do dynamic provisioning and volume binding once the PVC is created. You can set
it `WaitForFirstConsumer` to delay provisioning and binding until the PVC is being used.

#### A Use Case

This example utilizes dynamic provisioning to create storage for a zoo keeper service. (Here I simplify the config
for demo purpose. You can check [this doc] for more details about how to setup a Zoo Keeper Service with a StatefulSet.)

StatefulSet Spec:
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: zoo-keepr
# StatefulSet spec
spec:
  serviceName: zk-hs
  selector:
    matchLabels:
      app: zk
  replicas: 3

  volumeClaimTemplates:
  - metadata:
      name: datadir
    spec:
      storageClassName: standard
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 10Gi

  # Pod spec
  template:
    metadata:
      labels:
        app: zk
    spec:
      containers:
      - name: k8szk
        image: gcr.io/google_samples/k8szk:v3
        ...
        volumeMounts:
        - name: datadir
          mountPath: /var/lib/zookeeper
```

In this example, `volumeClaimTemplates` is used to do dynamic provisioning: A PVC is created with the specification defined
in `volumeClaimTemplates` for each Pod replica. Then A PV is created by `storageClass` `standard` with its specification.
 A 10Gi GCEPersistentDisk is created by the `storageClass` as the storage behind the PV. The PVC has the same
 `accessModes`, `storage` and `reclaimPolicy` with the PV.

## "Updating" PV & pvc (Create new one)

### Update Static PVs

Sometimes you need to update some parameters, for example, mount options, for your PV. This normally happens for static PVs,
which the storage behind your PV is not dynamically provisioned by any `StorageClass`. You can execute the following steps
to update your PV:
 1. Edit the PV, modify the parameters and save.
 2. Restart the Pods that are using the PV.

### Update Dynamic PVs

A PV now can be extended (shrinking is not supported) by editing its bound PVC in Kubernetes v1.11 or later versions. This feature is well supported
in many built-in volume provides, such as GCE-PD, AWS-EBS and GlusterFs. An cluster admin can make this feature available for cluster users by
setting `allowVolumeExpansion == true` in the `StorageClass`. Only PVCs created from the `StorageClass` can trigger volume expansion.
You can check [this blog](https://kubernetes.io/blog/2018/07/12/resizing-persistent-volumes-using-kubernetes/) for more details.
