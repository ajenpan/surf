#!/bin/bash
set -e # 遇到错误立即退出

latestbin="./battle"
runcmd="$latestbin"

echo "runcmd: $runcmd"

mkdir -p ./logs
chmod +x $latestbin

find_process_id() {
    echo $(ps -ef | grep $latestbin | grep -v grep | awk {'print $2'})
}

stop_process() {
    pids=$(find_process_id)
    if [ -n "$pids" ]; then
        echo "Sending SIGINT signal to process(es): $pids"
        kill -s SIGINT $pids
    fi

    timeout=300
    count=0
    while [ -n "$pids" ]; do
        sleep 100
        count=$((count + 100))
        if [ $count -eq $timeout ]; then
            break
        fi
    done

    if [ -n "$pids" ]; then
        echo "Process still running after $timeout seconds. Sending SIGKILL signal."
        kill -s SIGKILL $pids
    fi
}

stop_process

echo "start $runcmd"

# start process
nohup $runcmd 1>cout.log 2>cerr.log &

# sleep sceonds and check process stat
sleep 2s

# report
pids=$(find_process_id)
if [ -n "$pids" ]; then
    echo "$latestbin start success. processid: $pids"
    exit 0
else
    echo "$latestbin start failed"
    exit 1
fi
