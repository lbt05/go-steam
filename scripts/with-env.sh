#!/usr/bin/env bash
test -f .env && export $(grep -v '^#' .env | xargs -d '\n')
command "$@"
