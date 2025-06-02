#! /bin/bash

ft=""
mp="/mnt/fast-disks"
dp=""
isize=""
csize=""
nocheck=0
vgname="mlogvg"
while getopts ":t:i:c:m:d:g:n" opt; do
    case $opt in
        t)
            ft=$OPTARG
            ;;
        d)
            dp=$OPTARG
            ;;
        i)
            isize=$OPTARG
            ;;
        c)
            csize=$OPTARG
            ;;
        m)
            mp=$OPTARG
            ;;
        g)
            vgname=$OPTARG
            ;;
        n)
            nocheck=1
            ;;
        :)
            echo "wrong args"
    	    ;;
    esac
done


if [ "x${dp}" = "x" ]; then
  echo "Usage: $0 -d <disk_path>"
  exit 1
fi

if [[ "x${ft}" == "x" ]]; then
    un=`cat /etc/os-release | awk -F "=" '/^NAME/ {print $NF}'`
    un=${un//\"/}
    un=${un,,}
    if [ "${un}" = "ubuntu" ]; then
        ft="ext4"
    elif [[ "${un}" = "centos" || "${un}" = "redhat" ]]; then
	    ft="xfs"
    else
	    ft="ext4"
    fi
fi

if [[ "x${isize}" == "x" && "x${csize}" == "x" ]]; then
    fdiskl=`fdisk -l ${dp} | head -n 1 | awk '{print $5}'`
    fdiskl=$(( $fdiskl / 1024 / 1024 ))
    fsz=$(( $fdiskl - 16 ))
    isize="$(( $fsz / 10 ))m"
    csize="$(( $fsz / 10 * 9 ))m"
fi

if [[ $nocheck -eq 0 ]]; then
    lc=$(lsblk ${dp} | wc -l)
    if [[ ${lc} -gt 2 ]]; then
        echo "removing lvm first on ${dp}"
        exit -1
    fi
    if [[ ${lc} -eq 2 ]]; then
        lblk=`lsblk -p ${dp} | tail -n 1 | awk '{print $NF}'`
        if [[ "${lblk}" != "disk" ]]; then
            echo "removing lvm first on ${dp}"
            exit -1
        fi
    fi
fi


pvc=`pvcreate -t ${dp} 2>&1 | grep "successfully created\.$"`
if [[ "x${pvc}" == "x" ]]; then
    pvcreate ${dp}
fi
vgc=`vgcreate -t ${vgname} ${dp} 2>&1 | tail -n 1`
if [[ "${vgc}" =~ "successfully created" ]]; then
    vgcreate ${vgname} ${dp}
elif [[ "${vgc}" =~ "already exists" ]]; then
    echo ${vgc}
fi

# create index lvm
echo ${isize}
if [[ "x${isize}" != "x" ]]; then
    lvcreate -y -L ${isize} -n indexlvm ${vgname}
    mkfs.${ft} /dev/mapper/${vgname}-indexlvm
    sed -i "/\/dev\/mapper\/${vgname}-indexlvm/d" /etc/fstab
    echo "/dev/mapper/${vgname}-indexlvm ${mp}/indexlvm ${ft} defaults,noatime 0 0" >> /etc/fstab
fi


echo ${csize}
# create chunk lvm
if [[ "x${csize}" != "x" ]]; then
    lvcreate -y -L ${csize} -n chunklvm ${vgname}
    mkfs.${ft} /dev/mapper/${vgname}-chunklvm
    sed -i "/\/dev\/mapper\/${vgname}-chunklvm/d" /etc/fstab
    echo "/dev/mapper/${vgname}-chunklvm ${mp}/chunklvm ${ft} defaults,noatime 0 0" >> /etc/fstab
fi
