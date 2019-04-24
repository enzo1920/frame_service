FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install mc -y
RUN apt-get install net-tools -y

EXPOSE 8080:8080


RUN mkdir -p /home/frame_service
RUN mkdir -p /home/frame_service/log
RUN mkdir -p /home/frame_service/uploaded
RUN mkdir -p /home/frame_service/readconfig
COPY ./uploaded/Arduino_Switch03.png /home/frame_service/uploaded/Arduino_Switch03.png
COPY ./readconfig/frame_conf.conf /home/frame_service/readconfig/frame_conf.conf
COPY frameservice /home/frame_service/
WORKDIR /home/frame_service
ENTRYPOINT ["/home/frame_service/frameservice"]
#CMD ["/home/frame_service/frameservice"]