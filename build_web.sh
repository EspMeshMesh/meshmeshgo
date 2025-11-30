#!/bin/bash

cd web
npm run build
cd ..
git checkout managerui/managerui.go
