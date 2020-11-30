# Kubernetes Persistent Volumes and Persistent Volume Claims

## Persistent Volumes and Persistent Volume Claims

[Persistent Volumes (PVs)](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) and [Persistent Volume Claims (PVCs)](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) are designed for managing storage resources in Kubernetes.

The following picture shows the overview of PVs and PVCs.

![The Overview of Persistent Volumes and Persistent Volume Claims](https://github.com/azhuox/blogs/blob/master/kubernetes/pv_pvc/assets/pv-and-pvcs-overview.png?raw=true)


From the picture you can see that:

- PVs are created by cluster administrators and they are consumed by PVCs which are created by developers.
- A PV is like a mounting configuration of storage. Therefore, you can create different mount configurations for the same storage by creating multiple PVs.
- A PV is a public resource in a cluster, which means it is accessible to all the namespace. This also means the name of the PV needs to be unique in the whole cluster.
- A PVC is a k8s object within a namespace, which means its name must be unique in the namespace.
- A PV can only be exclusively bound to a PVC. This one-to-one mapping lasts until the PVC is deleted.
- A PV and its bound PVC builds a bridge between the "clients" (Pods) and the real storage.


## Provisioning Persistent Volumes

There are two ways to provision a PV: statically or dynamically.

### "Static" PVs

A static PV is a PV manually created by a cluster administrator with the details of a storage. 
"Static" here means the PV must exist before being consumed by a PVC.

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

In this example, the `ngx-pvc` PVC finds & binds the `nfs-pv` PV via [LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#labelselector-v1-meta). 
The `nfs-pod` Pod utilizes the `nfs-pvc` PVC to create a volume called `data-dir` and then mounts the volume to the `/usr/share/nginx/html` directory in the `nginx` container. 
After the creation of these Kubernetes objects, ssh into the `nginx` container in the `nfs-pod` Pod, run the `cat /proc/mount` command then you can find the information about the `nfs-pv` PV like: 
`12.34.56.78:/data /usr/share/nginx/html nfs4 vers=4.0,rsize=32768,wsize=32768,...,addr=12.34.56.78 0 0`. This means the NFS server `12.34.56.78:/data` which is specified in the `nfs-pv` PV is mounted to the `/usr/share/nginx/html` directory.

#### PV Types and Mount Options

Kubernetes currently supports many PV types, for example, NFS, CephFS, Glusterfs and GCEPersistentDisk. 
You can check [this doc](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#types-of-persistent-volumes) for more details.

In this example, the `nfs-pv` PV is created using NFS PV type, with the `12.34.56.78` server and the `data` path. In addition, this PV also specifies some other mount options for the NFS server. Mount options are only supported by some PV types, you can check [this doc](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#mount-options) for more details.

#### The capacity of A PV

The capacity of a static PV is not a hard limit of corresponding storage. Instead, the capacity is fully controlled by the real storage. 
Therefore, suppose the NFS server in the example has 200Gi storage space, the `nfs-pv` PV is able to use up all of the NFS server's space even although it only has `capacity.storage == 10Gi`. The capacity setting of a static PV normally is just for matching up the storage request in the corresponding PVC.

#### Access Mode

There are three access modes for a PV:

- `ReadWriteOnce`: a PV can be mounted as read-write by a single node if it has `ReadWriteOnce` in its `accessModes` spec. This means 1. the PV can perform read and write operation to storage. 2. The PV can only be mounted on a single node, which means any Pod that wants to use this PV must be scheduled to the same node as well.
- `ReadOnlyMany`: a PV can be mounted as read-only by many nodes if it has `ReadOnlyMany` in its `accessModes` spec. Unlike `ReadWriteOnce`, `ReadOnlyMany` allows the PV to be mounted on many nodes but it can only perform read operation to the real storage. Any write request will be denied in this case.
- `ReadWriteMany`: a PV can be mounted as read-write by many nodes if it has `ReadWriteMany` in its `accessModes` spec. This means the PV can perform read and write operations in many nodes.

Different PV types have different supports for these three access modes. You can check [this doc](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes) for more details.

You may notice that the PV's `accessModes` field is an array, which means it can have multiple access modes. **Nevertheless, a PV can only be mounted using one access mode at a time, even if it has multiple access mods in its `accessModes` field. Therefore, instead of including multiple access modes in a PV, it is recommended to have one access mode in one PV and create separate PVs with different access modes for different use cases.**

You may also notice that there are other attributes that can also affect access modes. Here is a simplified summary:

- The attribute `readOnly` of a PV type is a storage side setting. It is used to control whether real storage is read-only or not.
- `AccessModes` of a PV is a PV side setting and it is used to control the access mode of the PV.
- `AccessModes` of a PVC has to match up the PV that it wants to bind. 
A PV and a PVC build a bridge between the "client" and the real storage: the PV connects to the real storage while the PVC connects to the "client".
- The attribute `readOnly` of [VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#nfsvolumesource-v1-core) is a "client" side setting. It is used to control whether the mounted directory is read-only or not.

#### Reclaim Policy
The field `persistentVolumeReclaimPolicy` specifies the reclaim policy for a PV, which can be either `Delete` (default value) or `Retain`. You may want to set it `Retain` for a PV and back up the data at a certain frequency if the data inside the storage that the PV connects is really important.

#### Binding

The example above uses [LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#labelselector-v1-meta) `matchLabels.pv-name == pv-name` to bind the `nfs-pv` PV and the `nfs-pvc` PVC together. You do not need to use LabelSelector to establish the bind between PVs and PVCs if you want a more flexible way of binding. For example, without LabelSelector, a PVC that requires `storage == 10Gi` and `accessModes == [ReadWriteOnce]` can be bound to a PV with `storage >= 10Gi` and `accessModes == [ReadWriteOnce, ReadWriteMany]`.


### "Dynamic" Persistent Volumes

Dynamic PVs are dynamically created by K8s, which is triggered by the specification of a user's PVC. The dynamic provisioning is based on [Storage Classes](https://kubernetes.io/docs/concepts/storage/storage-classes/): a PVC must specify an existing `StorageClass` in order to create a dynamic PV.

#### Storage Classes.

A `StorageClass` is a Kubernetes object used to describe a storage class. It uses fields like `parameters`, `provisioner` and `reclaimPolicy` to describe details of the storage class that it represents. 

Let's take a look at the GKE's default storage class `standard`, here is its spec:

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
- The field `parameters` specifies the parameters for the real storage. For example, `parameters.type == pd-standard` means this storage class uses [GCEPersistentDisk](https://kubernetes.io/docs/concepts/storage/volumes/#gcepersistentdisk) as storage media. You can check [this doc](https://kubernetes.io/docs/concepts/storage/storage-classes/#parameters) for more details about the parameters of Storage Classes.
- The field `provisioner` specifies which volume plugin is used by the Storage Class to provision dynamic PVs. You can check [this list](https://kubernetes.io/docs/concepts/storage/storage-classes/#provisioner) for each provisioner's specification.
- Like field `persistentVolumeReclaimPolicy`, the field `reclaimPolicy`  specifies the reclaim policy for the storage created by the Storage Class. It can be either `Delete` (default value) or `Retain`.
- The field `volumeBindingMode` controls when to perform dynamic provisioning and volume binding. `volumeBindingMode == Immediate` means doing dynamic provisioning and volume binding once the PVC is created, while `volumeBindingMode == WaitForFirstConsumer` means delaying dynamic provisioning and volume binding until the PVC is actually being consumed.

#### A Use Case

This example utilizes dynamic provisioning to create storage resources for a ZooKeeper service. (Here I simplify the config for the demo purpose. 
You can check [this doc](https://github.com/kubernetes/contrib/blob/master/statefulsets/zookeeper/zookeeper.yaml) for more details about how to set up a ZooKeeper Service with a StatefulSet in Kubernetes.)

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

In this example, the field `volumeClaimTemplates` is used to do dynamic provisioning: A PVC is created with the storage specification defined in the `volumeClaimTemplates` field for each Pod. Then A PV is created by the `standard` Storage Class for each Pod and bound to each Pod's PVC. Then A 10Gi GCEPersistentDisk is created by the `standard` Storage Class for each PV. A PVC has the same `accessModes`, `storage` and `reclaimPolicy` with its corresponding PV.

## "Updating" PVs & PVCs

### Updating Static PVs

Sometimes you need to update some parameters, for example, mount options, for a static PV, in which the storage is not dynamically provisioned. 
However, updating a PV that is being used may be blocked by K8s. But as I mentioned above, A PV and PVC bound is like building a bridge between clients and the real storage. 
Therefore, instead of updating the existing PV, you can create a new PV and PVC with the new settings you want, and then mount replace the old PVC with the new one.

### Updating Dynamic PVs

A dynamic PV now can be extended (shrinking is not supported) by editing its bound PVC in Kubernetes v1.11 or later versions. 
This feature is well supported in many built-in volume providers, such as GCE-PD, AWS-EBS and GlusterFs. 
A cluster administrator can make this feature available for cluster users by setting `allowVolumeExpansion == true` in the configurations of the Storage Classes. 
You can check [this blog](https://kubernetes.io/blog/2018/07/12/resizing-persistent-volumes-using-kubernetes/) for more details.

## What Is Next

This is the last blog of my series of blogs about the introduction to Kubernetes. Now you should have brief idea about what is Kubernetes.
I highly recommend you check [the official Kubernetes documentation](https://kubernetes.io/docs/home/) if you want to dive deeply into Kubernetes.

## Reference

- [Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [Persistent Volume Claims](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims)
- [LabelSelector in Kubernetes](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#labelselector-v1-meta)
- [Storage Classes](https://kubernetes.io/docs/concepts/storage/storage-classes/)
- [The StatefulSet Configuration for Setting Up A ZooKeeper Service](https://github.com/kubernetes/contrib/blob/master/statefulsets/zookeeper/zookeeper.yaml)
- [Resizing Persistent Volumes using Kubernetes](https://kubernetes.io/blog/2018/07/12/resizing-persistent-volumes-using-kubernetes/)
