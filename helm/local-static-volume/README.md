# Deploy Local Static Volume

It is recommended that all loki-write nodes contain identical block device to store index/chunk, and use separate lvms to store index/chunk(5%-15% for index). detail as below(ubuntu 22.04):

## Setup local static volume step by step

1. mount block device(with lvm)
```
root@k8s190:/root# pvcreate /dev/sdb
root@k8s190:/root# vgcreate mlogvg /dev/sdb
root@k8s190:/root# lvcreate -n indexlvm --size 5g mlogvg
root@k8s190:/root# lvcreate -n chunklvm --size 24g mlogvg
################################################################
# Note: both indexlvm and chunklvm are used by loki, indexlvm store indexes, and chunklvm store log segments, detail in https://grafana.com/docs/loki/latest/operations/storage/, and these two values will be used during install loki, using a fixed value is a good choice
################################################################
root@k8s190:/root# mkdir -p /mnt/fast-disks/{indexlvm,chunklvm}
root@k8s190:/root# mkfs -t ext4 /dev/mapper/mlogvg-indexlvm
root@k8s190:/root# echo "/dev/mapper/mlogvg-indexlvm /mnt/fast-disks/indexlvm ext4 defaults,noatime 0 0" >> /etc/fstab
root@k8s190:/root# mount /mnt/fast-disks/indexlvm

root@k8s190:/root# mkfs -t ext4 /dev/mapper/mlogvg-chuncklvm
root@k8s190:/root# echo "/dev/mapper/mlogvg-chunklvm /mnt/fast-disks/chunklvm ext4 defaults,noatime 0 0" >> /etc/fstab
root@k8s190:/root# mount /mnt/fast-disks/chunklvm
```

2. add labels for those k8s nodes which used to running loki components(write)
```
root@k8s190:/root# kubectl label node k8s190 aplication/component=loki-write
root@k8s190:/root# kubectl label node k8s191 aplication/component=loki-write
root@k8s190:/root# kubectl label node k8s192 aplication/component=loki-write
```
3. edit the value.yaml for install sig-storage-local with helm
```
root@k8s190:/root# cat << EOF > vol-values.yaml
classes:
  - name: fast-disks
    hostDir: /mnt/fast-disks  # the parent folder of the mount point
    fsType: ext4 # same with "mkfs -t ext4 ..."
nodeSelector:
  aplication/component: loki-write
image: registry.k8s.io/sig-storage/local-volume-provisioner:v2.7.0 #crpi-2re4a582sqaza89h.cn-hangzhou.personal.cr.aliyuncs.com/venomous/local-volume-provisioner:v2.7.0
EOF
root@k8s190:/root# help install --values ./vol-values.yaml sig-local-static-volume ./provisioner
```
Note: If each write node uses a block device of different capacity, use the --size parameter to create indexlvm and chunklvm logical volumes and ensure that each node has the same capacity.  
4. display pv(s) in k8s cluster
```
root@k8s190:/root# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                 STORAGECLASS   REASON   AGE
local-pv-50fd5c82                          24Gi       RWO            Delete           Available                         fast-disks              
local-pv-6d92d7b1                          4449Mi     RWO            Delete           Available                         fast-disks              
local-pv-8cf428dc                          4449Mi     RWO            Delete           Available                         fast-disks              
local-pv-95dc1911                          4449Mi     RWO            Delete           Available                         fast-disks              
local-pv-98724ea2                          24Gi       RWO            Delete           Available                         fast-disks              
local-pv-9fd2def3                          24Gi       RWO            Delete           Available                         fast-disks              
root@k8s190:/root# 
```
5. tips(createlvm.sh to create lvm)
```
createlvm.sh usage
-d block device, e.g. -d /dev/sdb
-t fs type, default: ext4 for ubuntu and xfs for centos
-m mount point, default: /mnt/fast-disks
-g vg-name, default: mlogvg
-i index-lvm size, default: 10 percent of the block device
-c chunk-lvm size, default: 90 percent of the block device
```