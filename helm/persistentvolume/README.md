# Deploy Local Static Volume

It is recommended to use a separate block device for a local static volume in kubernetes. detail as below(ubuntu 22.04):

## Setup local static volume step by step

1. mount block device(lvm)
```
root@minlog01:/root# pvcreate /dev/sdb
root@minlog01:/root# vgcreate k8svg /dev/sdb
root@minlog01:/root# lvcreate -n k8slv -l 100%FREE k8svg
root@minlog01:/root# duid="dm-uuid-LVM-$(vgdisplay k8svg | awk '/VG UUID/{gsub(/-/,"",$NF);print $NF}')$(lvdisplay /dev/k8svg/k8slv | awk '/LV UUID/{gsub(/-/,"",$NF);print $NF}')"
root@minlog01:/root# mkdir -p /mnt/fast-disks/lv-minlog
root@minlog01:/root# mkfs -t ext4 /dev/disk/by-id/${duid}
root@minlog01:/root# echo "/dev/disk/by-id/${duid} /mnt/fast-disks/lv-minlog ext4 defaults 0 1" >> /etc/fstab
root@minlog01:/root# mount /mnt/fast-disks/lv-minlog
``` 