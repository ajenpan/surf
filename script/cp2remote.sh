#!/bin/bash

set -e # 如果任何命令失败，脚本将立即退出

BinFile=$1

if [ -z "$BinFile" ]; then
    echo "Usage: $0 <binfile>"
    exit 1
fi

ScriptDir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

NowTime=$(date +%Y%m%d%H%M%S)
TargetBin=$(realpath "${ScriptDir}/../bin/${BinFile}")

echo "currdir:$(pwd), ScriptDir: ${ScriptDir}, TargetBin: ${TargetBin}"

BinDir=/workdir/server/${BinFile}
RemoteHost=root@myali01

echo "start to upload ${TargetBin} to ${RemoteHost}:${BinDir}/${BinFile}.tmp"

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

echo "deploy ${BinFile} to ${RemoteHost} done!"
