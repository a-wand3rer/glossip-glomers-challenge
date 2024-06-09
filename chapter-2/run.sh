#!/bin/bash

cwd=$(pwd)
maelstrom_dir=$HOME/bin/maelstrom
go build -o build/node-unique-ids
cd $maelstrom_dir
./maelstrom test -w unique-ids --bin $cwd/build/node-unique-ids --node-count 3 --time-limit 30 --rate 1000 --availability total --nemesis partition
cd $cwd
