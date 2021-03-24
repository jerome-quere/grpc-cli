#!/bin/bash

protoc -I ./protobuf $(find ./protobuf -name "*.proto") -o ./internal/args/testdata/test.pb