#!/usr/bin/env bash

go run server/main.go &
go run balancer/main.go
