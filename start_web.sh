#!/bin/bash

cdir=`dirname $0`
cd $cdir

PID_FILE=/tmp/gopush-cluster-web.pid

./web  -v=3  -log_dir="../logs/web/" -stderrthreshold=FATAL &

echo $! > $PID_FILE

wait

rm -f $PID_FILE > /dev/nulli

cd -
