#!/bin/bash
#

sed -i 's/package main/package meminject/g' *.go
sed -i 's/*verbose/verbose/g' *.go
sed -i 's/*debug/debug/g' *.go
sed -i 's/*pid/pid/g' *.go
sed -i 's/*program/program/g' *.go
sed -i 's/*args/args/g' *.go
