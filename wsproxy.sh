#!/bin/bash

docker run --rm -v $PWD/nginx.conf:/etc/nginx/conf.d/default.conf --name nginx -p 80:80 nginx:latest