#! /bin/sh

hn=${HOSTNAME}
lep=${LOKIEP:-http://loki:3100/loki/api/v1/push}
ro=${RUNNINGONLY:-true}
scmd="run"
alloyfile=${ALLOYFILE:-/app/confs/config.alloy}
ns=${NAMESPACES}


while getopts ":r:l:h:c:f:n:" opt; do
    case $opt in
        c)
            scmd=$OPTARG
            ;;
        r)
            ro=$OPTARG
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
    margs="--namespace ${ns} ${margs}"
fi
if [[ "x${hn}" != "x" ]]; then
    margs="--hostname ${hn} ${margs}"
fi

if [[ "${ro,,}" == "true" || "${ro,,}" == "yes" ]]; then
    margs="${margs} --running-only"
fi

/app/minlog ${scmd} ${margs}

