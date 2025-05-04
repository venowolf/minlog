#! /bin/sh

hn=${HOSTNAME}
lep=${LOKIEP:-http://loki:3100/loki/api/v1/push}
ro=${RUNNINGONLY:-true}

while getopts ":r:l:h:" opt; do
    case $opt in
        r)
            ro=$OPTARG
            ;;
        l)
            lep=$OPTARG
            ;;
        h)
            hn=$OPTARG
            ;;
	:)
            echo "wrong args"
    	    ;;
    esac
done

margs="--lokiep ${lep}"
if [[ "x${hn}" != "x" ]]; then
    margs="--hostname ${hn} ${margs}"
fi

if [[ "${ro,,}" == "true" || "${ro,,}" == "yes" ]]; then
    margs="${margs} --running-only"
fi

/app/minlog run ${margs}

