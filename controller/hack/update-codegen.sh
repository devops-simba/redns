#!/bin/sh

if [ -z "$GOPATH" ]; then
  GOPATH="$HOME/go"
fi

SCRIPT_PATH="$0"
SCRIPT_FOLDER="`dirname $SCRIPT_PATH`"
BOILERPLATE_FILE="$SCRIPT_FOLDER/boilerplate.go.txt"
ROOT_PACKAGE="github.com/devops-simba/redns/controller"
OUTPUT_PATH="$GOPATH/src"

echo "OUTPUT_PATH: $OUTPUT_PATH"
$GOPATH/src/k8s.io/code-generator/generate-groups.sh all \
  "$ROOT_PACKAGE/pkg/client" \
  "$ROOT_PACKAGE/pkg/apis" \
  "redns:v1" \
  --output-base "$OUTPUT_PATH" \
  --go-header-file "$BOILERPLATE_FILE"
