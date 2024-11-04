#!/bin/bash

set -e # 如果任何命令失败，脚本将立即退出

BinFile=$1

if [ -z "$BinFile" ]; then
    echo "Usage: $0 <binfile>"
    exit 1
fi

NowTime=$(date +%Y%m%d%H%M%S)

TargetBin=../bin/${BinFile}

BinDir=/root/svr_run/${BinFile}/
RemoteHost=root@myali01

scp ${TargetBin} ${RemoteHost}:${BinDir}/${BinFile}.tmp

ssh ${RemoteHost} >/dev/null 2>&1 <<EOF
cd ${BinDir}

if [ -f ${BinFile} ]; then
    mv ${BinFile} ${BinFile}.${NowTime}
fi

mv ${BinFile}.tmp ${BinFile}

./run.sh

exit
EOF

echo "Deployment ${BinFile} to ${RemoteHost} done!"
