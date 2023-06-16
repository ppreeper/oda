#!/usr/bin/env bash
while inotifywait -e create,delete,modify -m -r src/ ; do
  bashly generate
done
