#!/bin/bash

cdir=`dirname $0`
cd $cdir

nohup ./web  -v=3 -log_dir="../logs/web/" -stderrthreshold=INFO &
nohup ./message  -v=1 -log_dir="../logs/message/" -stderrthreshold=FATAL &
nohup ./comet  -v=1 -log_dir="../logs/comet/" -stderrthreshold=FATAL &

cd -
