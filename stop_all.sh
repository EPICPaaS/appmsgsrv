#!/bin/bash

cdir=`dirname $0`
cd $cdir

kill `cat /tmp/gopush-cluster-comet.pid`
kill `cat /tmp/gopush-cluster-message.pid`
kill `cat /tmp/gopush-cluster-web.pid`

cd -
