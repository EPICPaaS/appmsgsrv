#!/bin/bash

cdir=`dirname $0`
cd $cdir

PID_FILE=/tmp/gopush-cluster-comet.pid

./comet  -v=1 -log_dir="../logs/comet/" -stderrthreshold=FATAL  &

echo $! > $PID_FILE

wait

rm -f $PID_FILE > /dev/nulli

cd -
