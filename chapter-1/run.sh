#!/bin/bash

cwd=$(pwd)
maelstrom_dir=$HOME/bin/maelstrom
go build -o build/app
cd $maelstrom_dir
./maelstrom test -w echo --bin $cwd/build/app --node-count 1 --time-limit 10
cd $cwd
