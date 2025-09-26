#!/bin/bash
set -euxo pipefail

# Define the URL and options
url="http://localhost:8080/vote"
options=("Hello" "World")

# Loop through each option and send the requests
for option in "${options[@]}"; do
    for i in {1..3}; do
        curl -X POST -H "Content-Type: application/json" -d "{\"option\": \"$option\"}" "$url"
    done
done

# Get the results
curl -X GET http://localhost:8081/results | jq .
