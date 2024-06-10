#!/bin/bash

cwd=$(pwd)
maelstrom_dir=$HOME/bin/maelstrom
rm build/node-multi-broadcast
echo ">> build new binary"
go build -o build/node-multi-broadcast
$maelstrom_dir/maelstrom test -w broadcast --bin $cwd/build/node-multi-broadcast --node-count 5 --time-limit 20 --rate 10
cd $cwd
