#!/bin/zsh
date
{
  cmd="go run hello.go"
  eval "${cmd}"
} &
sleep 2
{
  cmd="go run gateway.go"
  eval "${cmd}"
} &

wait
date