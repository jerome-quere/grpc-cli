#!/bin/bash

protoc -I ./protobuf ./protobuf/test/test.proto -o ./internal/args/testdata/test.pb