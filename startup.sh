#! /bin/bash

hn=${HOSTNAME}
lep=${LOKIEP:-http://loki:3100/loki/api/v1/push}
ro=${RUNNINGONLY:-true}
scmd="run"
alloyfile=${ALLOYFILE:-/etc/alloy/alloy.alloy}
ns=${NAMESPACES}
incluster=0


while getopts ":rl:h:c:f:n:k" opt; do
    case $opt in
        c)
            scmd=$OPTARG
            ;;
        r)
            ro="true"
            ;;
        l)
            lep=$OPTARG
            ;;
        h)
            hn=$OPTARG
            ;;
        f)
            alloyfile=$OPTARG
            ;;
        k)
            incluster=1
            ;;
        n)
            ns=$OPTARG
            ;;
        :)
            echo "wrong args"
    	    ;;
    esac
done

margs="--lokiep ${lep} --alloy-file ${alloyfile}"
if [[ "x${ns}" != "x" ]]; then
    margs="--namespaces ${ns} ${margs}"
fi
if [[ "x${hn}" != "x" ]]; then
    margs="--label-nodename ${hn} ${margs}"
fi

if [[ "${ro,,}" == "true" || "${ro,,}" == "yes" ]]; then
    margs="${margs} --running-only"
fi
if [[ ${incluster} -eq 1 ]]; then
    margs="${margs} --running-in-cluster"
fi
echo "running: /app/minlog ${scmd} ${margs}"
/app/minlog ${scmd} ${margs}

