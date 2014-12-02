#!/bin/bash

cdir=`dirname $0`
cd $cdir

PID_FILE=/tmp/gopush-cluster-message.pid

./message  -v=1 -log_dir="../logs/message/" -stderrthreshold=FATAL &

echo $! > $PID_FILE

wait

rm -f $PID_FILE > /dev/nulli

cd -
