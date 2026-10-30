#!/bin/zsh

for (( i = 0; i < 100; i++ )); do
    curl http://localhost:8080/hello
done