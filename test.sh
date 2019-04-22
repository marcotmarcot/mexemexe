#!/bin/sh

set -e
set -u
set -o pipefail

for i in 1 2 3 4 5 real1; do
    echo Running $i
    cat $i.in | go run mexemexe2.go 2>/dev/null | diff - $i.out && echo Success || echo Failure
done
