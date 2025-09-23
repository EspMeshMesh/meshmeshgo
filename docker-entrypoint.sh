#!/bin/sh
DELAY=${MESHMESH_RESTART_DELAY:-5}   # default to 5 seconds if not set

while true; do
  ./meshmeshgo
  echo "meshmeshgo crashed (exit code $?). Restarting in ${DELAY}s..."
  sleep "$DELAY"
done
