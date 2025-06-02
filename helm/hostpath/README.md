# store loki indexes and chunks with hostPath

## planning
It is a good idea to deploy loki using independent nodes that meet the I/O, RAM, CPU, and Network requirements, and store loki indexes and chunks in /data/loki, e.g. loki01, loki02, loki03
```
root@k8s190:~# kubectl get nodes
NAME     STATUS   ROLES           AGE   VERSION
loki01   Ready    <none>          32d   v1.28.2
loki02   Ready    <none>          60d   v1.28.2
loki03   Ready    <none>          60d   v1.28.2
......
```

## Setup
1. Create labels for loki nodes
```
root@k8s190:~# kubectl label node lokie01 loki.component=write
root@k8s190:~# kubectl label node lokie02 loki.component=write
root@k8s190:~# kubectl label node lokie03 loki.component=write
```
2. Copy and edit helm value( hostpath-values-template.yaml )
```
root@k8s190:~# cat >> hostpath-values.yaml < EOF

EOF
```
3.  Create Local Volume(using /loki-data to store loki persistant data, including indexes chunks and other data that needs to be stored locally)
4.