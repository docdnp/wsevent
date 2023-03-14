#!/bin/bash

docker run --rm -v $PWD/nginx.conf:/etc/nginx/conf.d/default.conf --network host --name nginx -p 80:80 nginx:latest

docker ps | grep -Eo 'wsclient.*'  | xargs docker kill

