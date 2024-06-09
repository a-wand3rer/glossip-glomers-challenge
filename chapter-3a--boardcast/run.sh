#!/bin/bash

cwd=$(pwd)
maelstrom_dir=$HOME/bin/maelstrom
go build -o build/node-broadcast
$maelstrom_dir/maelstrom test -w broadcast --bin $cwd/build/node-broadcast --node-count 1 --time-limit 20 --rate 10
cd $cwd
