#!/bin/bash
set -eu
tps=$1
rm access.log
./nginx-random-access-log-generator -tps $tps &
sleep 10
kill $!
echo "tps=$tps $(wc -l access.log)"
