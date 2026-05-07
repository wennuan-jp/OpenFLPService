#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: ./test_api.sh <path_to_flp_file>"
    exit 1
fi

FLP_FILE=$1

echo "Testing FLP upload to http://localhost:8080/upload..."
curl -X POST -F "file=@$FLP_FILE" http://localhost:8080/upload | jq .
