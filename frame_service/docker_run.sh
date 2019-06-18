#!/bin/sh
docker run -dit --name framer2 -p 8080:8080 -v /home/go_projects/src/frame_service/front:/home/frame_service/front   -v /home/go_projects/src/frame_service/uploaded:/home/frame_service/uploaded bc231f3f0243
