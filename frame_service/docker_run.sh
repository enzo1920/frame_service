#!/bin/sh
docker run -dit --name framer2 -p 80:80 -v /home/go_projects/src/frame_service/front:/home/frame_service/front   -v /home/go_projects/src/frame_service/uploaded:/home/frame_service/uploaded 59b979a968ba
