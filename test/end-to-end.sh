#!/bin/bash

set -a # auto export variables
set -e # exists on non-zero errors
set -u # fail on undefined vars
# set -v # verbose print of commands before execution

# Laravel create/preset

kool create laravel test_laravel_app

echo "Asserting tests..."

# assert files exist
[ -d test_laravel_app/ ] || (echo "failed creating laravel app folder" && exit 2)
[ -f test_laravel_app/docker-compose.yml ] || (echo "failed installing preset docker-compose.yml on laravel" && exit 2)
[ -f test_laravel_app/kool.yml ] || (echo "failed installing preset kool.yml on laravel" && exit 2)

cd test_laravel_app
cp .env.example .env
kool start

# check containers are running
[ $(kool status | grep 'Running' | wc -l | grep 3) ] || (echo "bad count for running containers" && exit 2)

# check default PHP 7.4
[ $(kool exec app php -v | grep 'PHP 7.4' | wc -l | grep 1) ] || (echo "bad version" && exit 2)

cd ..
rm -rf test_laravel_app/
