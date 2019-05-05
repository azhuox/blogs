# Persistent Volumes and Persistent Volume Claims in Kubernetes

What is included in this blog:

- An introduction of Persistent Volumes and Persistent Volume Claims.
- A Discussion about how to update Persistent Volumes and Persistent Volume Claims.


# What are Persistent Volumes and Persistent Volume Claims

[Persistent Volumes (PVs)](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) and [Persistent Volume Claims (PVCs)](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) are designed for managing storage resources in Kubernetes.

The following picture shows the overview of PVs and PVCs.

[image]

From the picture you can see that:

- PVs are created by cluster admins and they are consumed by PVCs which are created by developers.
- A PV is like a mount configuration to a storage. Therefore, you can create different mount configurations for the same storage by creating multiple PVs.
- A PV is a public resource in a cluster, which means it is accessible to all the namespace. This also means the name of the PV needs to be unique in the whole cluster.
- A PVC is a k8s object within a namespace. Its name must be unique in the namespace.
- A PV can only be exclusively bound to a PVC. This one-to-one mapping lasts until the PVC is deleted.


# Provisioning

There are two ways to provision PVs: statically or dynamically.

## "Static" PVs

A static PV means it is manually created by a cluster administrator with details of a storage. "Static" means the PV must exist before being consumed by a PVC.

Here is an example of static PVs:

PV spec:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-pv
spec:
  nfs:
    # TODO: use right IP
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

In this example, the `nfs-pod` Pod utilizes the `nfs-pvc` PVC to create a volume called `data-dir` and then mounts the volume to the directory `/usr/share/nginx/html` in the container `nginx`. The PVC `ngx-pvc` finds and binds the `nfs-pv` PV via [LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#labelselector-v1-meta). After the creation of these Kubernetes objects, ssh into the `nginx` container in the `nfs-pod` Pod and run `cat /proc/mount` then you will find the mount information of the `nfs-pv` PV, like: `12.34.56.78:/data /usr/share/nginx/html nfs4 vers=4.0,rsize=32768,wsize=32768,...,addr=12.34.56.78 0 0`. This means the directory `/usr/share/nginx/html` is mapped to `12.34.56.78:/data`.

### PV Types and Mount Options
Kubernetes currently supports a lot of PV types, for example NFS, CephFS, Glusterfs and GCEPersistentDisk. You can check [this doc](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#types-of-persistent-volumes) for more details.

In this example, the `nfs-pv` PV is created using NFS PV type, with server `12.34.56.78` and the path `data`. In addition, this PV also specifies some other mount options for connecting to the NFS server. Mount options are only supported by some PV types, you can check [this doc](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#mount-options) for more details.

### The capacity of A PV
The capacity of a static PV is not hard limit of the corresponding storage. Instead, the capacity is fully controlled by the real storage. Therefore, suppose the NFS server in the example has 200Gi storage space, the `nfs-pv` PV is able to use up all of the NFS server's space even although it only has `capacity.storage == 10Gi`. Additionally, The capacity setting of a static PV normally is just for matching up the storage request in the corresponding PVC.

### Access Mode
There are three access modes for PVs and PVCs:

- `ReadWriteOnce`: a PV can be mounted as read-write by a single node if it has `ReadWriteOnce` in its `accessModes` spec. This means 1. the PV can perform read and write operation to a storage. 2. The PV can only be mounted on a single node, which means any Pod that wants to use this PV must be scheduled to the same node as well.
- `ReadOnlyMany`: a PV can be mounted as read-only by many nodes if it has `ReadOnlyMany` in its `accessModes` spec. Unlike `ReadWriteOnce`, `ReadOnlyMany` allows the PV to be mounted on many nodes although it can only perform read to real storage. Any write request will be denied in this case.
- `ReadWriteMany`: a PV can be mounted as read-write by many nodes if it has `ReadWriteMany` in its `accessModes` spec.

Different PV types have different supports for these three accessModes. You can check [this doc](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes) for more details.

You may notice that the PV's `accessModes` field is an array, which means it can has multiple accessModes. **Nevertheless, a PV can only be mounted using one access mode at a time, even if it has multiple access mods in its `accessModes` field. Therefore, instead of including multiple access modes in a PV, it is recommended to have one access mode in one PV and create separate PVs with different access modes for different use cases.**

You may also notice that there are other attributes which can also affect access modes. Here is simplified summary:

- `readOnly` attribute of a PV type is storage side setting. It is used to control whether real storage is read-only or not.
- `AccessModes` of a PV is PV side setting and it is used to control access mode of the PV.
- `AccessModes` of a PVC has to match up the PV that it wants to bind. The PV and PVC build a bridge between the "client" and the real storage: PV connects to the real storage while PVC connects to the "client".
- `readOnly` attribute of [VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#nfsvolumesource-v1-core) is "client" side setting. It is used to control whether the mounted directory is read-only or not.

### Reclaim Policy
The field `persistentVolumeReclaimPolicy` specifies reclaim policy for the PV, which can be either `Delete` (default value) or `Retain`. You may want to set it `Retain` and back up data in the PV yourself if the data inside the storage that PV connects to is really important.

### Binding
The example above uses [LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#labelselector-v1-meta) `matchLabels.pv-name == pv-name` to bind the `nfs-pv` PV and the `nfs-pvc` PVC together. You do not need to use LabelSelector to establish the bind between PVs and PVCs if you want more flexible way of binding. For example, without LabelSelector, a PVC which requires `storage == 10Gi` and `accessModes == [ReadWriteOnce]` can be bound to a PV with `storage >= 10Gi` and `accessModes == [ReadWriteOnce, ReadWriteMany]`.


## "Dynamic" Persistent Volumes

Dynamic Persistent Volumes are the volumes which are dynamically created by the cluster with specification of a user's PVC. This provisioning is based on [Storage Classes](https://kubernetes.io/docs/concepts/storage/storage-classes/): the PVC must specify an existing `StorageClass` in order to create a dynamic PV.

### Storage Classes.

A `StorageClass` is a Kubernetes object used to describe a class of storage. It uses the fields like `parameters`, `provisioner` and `reclaimPolicy` to describe details of the storage class that it represents. Let's take a look at the GKE's default storage class `standard`, here is its spec:

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

Explanation:

- The field `metadata.name` is the name of the `StorageClass`. It has to be unique in the whole cluster.
- The field `parameters` specifies parameters for real storage. For example `parameters.type == pd-standard` means this storage class uses [GCEPersistentDisk](https://kubernetes.io/docs/concepts/storage/volumes/#gcepersistentdisk) as storage media. You can check [this doc](https://kubernetes.io/docs/concepts/storage/storage-classes/#parameters) for more details about the parameters of Storage Classes.
- The field `provisioner` specifies which volume plugin is used for provisioning dynamic PVs for the Storage Class. You can check [this list](https://kubernetes.io/docs/concepts/storage/storage-classes/#provisioner) for each provisioner's specification.
- Like `persistentVolumeReclaimPolicy`, the field `reclaimPolicy` specifies the reclaim policy for the storage created by the Storage Class, which can be either `Delete` (default value) or `Retain`.
- The field `volumeBindingMode` controls when to do dynamic provisioning for the PV and volume binding. `volumeBindingMode == Immediate` means doing dynamic provisioning and volume binding once the PVC is created. You can set it `volumeBindingMode == WaitForFirstConsumer` to delay dynamic provisioning and volume binding until the PVC is being firstly comsued.

### A Use Case

This example utilizes dynamic provisioning to create storage for a ZooKeeper service. (Here I simplify the the config for the demo purpose. You can check [this doc](https://github.com/kubernetes/contrib/blob/master/statefulsets/zookeeper/zookeeper.yaml) for more details about how to setup a ZooKeeper Service with a StatefulSet in Kubernetes.)

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

In this example, `volumeClaimTemplates` is used to do dynamic provisioning: A PVC is created with the specification defined in `volumeClaimTemplates` for each Pod. Then A PV is created by the `standard` Storage Class. Then A 10Gi GCEPersistentDisk is created by the `standard` Storage Class for each PV. The PVC has the same `accessModes`, `storage` and `reclaimPolicy` with the PV.

# "Updating" PVs & PVCs

### Updating Static PVs

Sometimes you need to update some parameters, for example, mount options, for a static PV which the storage is not dynamically provisioned. You can execute the following steps to update a static PV:
 
- 1. Edit the PV, modify the parameters and save.
- 2. Restart the Pods that are using the PV.

### Updating Dynamic PVs

A dynamic PV now can be extended (shrinking is not supported) by editing its bound PVC in Kubernetes v1.11 or later versions. This feature is well supported in many built-in volume providers, such as GCE-PD, AWS-EBS and GlusterFs. An cluster admin can make this feature available for cluster users by setting `allowVolumeExpansion == true` in the configurations of the Storage Classes. You can check [this blog](https://kubernetes.io/blog/2018/07/12/resizing-persistent-volumes-using-kubernetes/) for more details.

# Reference

- [Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [Persistent Volume Claims](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims)
- [LabelSelector in Kubernetes](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#labelselector-v1-meta)
- [Storage Classes](https://kubernetes.io/docs/concepts/storage/storage-classes/)
- [The StatefulSet Configuration for Setting Up A ZooKeeper Service](https://github.com/kubernetes/contrib/blob/master/statefulsets/zookeeper/zookeeper.yaml)
- [Resizing Persistent Volumes using Kubernetes](https://kubernetes.io/blog/2018/07/12/resizing-persistent-volumes-using-kubernetes/)
