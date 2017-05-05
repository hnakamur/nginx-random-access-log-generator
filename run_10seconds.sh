#!/bin/bash
set -eu
rm -f access.log
./nginx-random-access-log-generator
