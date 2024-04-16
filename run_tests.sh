#!/bin/bash

EXCLUDE_DIRS="adapters/ent"

CMD="go test -coverprofile=coverage.out"

for PKG in $(go list ./... | grep -Ev "$(echo $EXCLUDE_DIRS | sed 's/ /|/g')"); do
    CMD+=" $PKG"
done

echo "Running command: $CMD"
eval $CMD