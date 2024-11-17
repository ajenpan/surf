#!/bin/bash

protoc --version

pbfiles=$(find ../proto -name "*.proto" -type f)

# todo:
# for file in $pbfiles; do
#     filename=$(basename -- "$file")
#     dir=$(dirname "$file")
#     echo $dir $filename $file
#     protoc -I=${dir} -I=. --go_out=../msg $file
# done
