#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

go_bin=go

targets=("ralph")

function check() {
    if [ $# -ne 2 ]; then
        echo "Usage: check <array> <value>"
        return 1
    fi

    local array_name="$1"
    local value="$2"
    local array=()

    if [ "$array_name" == "targets" ]; then
        array=("${targets[@]}")
    else
        echo "Unknown array name: $array_name"
        return 1
    fi

    for e in "${array[@]}"; do
        if [ "$e" == "$value" ]; then
            return 0
        fi
    done

    echo "check failed, $value not in ${array[*]}"
    exit 1
}

target="${targets[0]}"

while [ "$#" -gt 0 ]; do
    case $1 in
    -t | --target)
        target="$2"
        check "targets" "$target"
        if [ $? -ne 0 ]; then
            echo "target must be one of ${targets[*]}"
            exit 1
        fi
        shift 2
        ;;
    *)
        echo "unknown command $1"
        exit 1
        ;;
    esac
done

echo "building target: $target"

rm -rf output
mkdir -p output

$go_bin build -o "output/$target" "./cmd/$target"
