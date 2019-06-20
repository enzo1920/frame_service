FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install mc -y
#RUN apt-get install net-tools -y

EXPOSE 80:80


RUN mkdir -p /home/frame_service
RUN mkdir -p /home/frame_service/log
RUN mkdir -p /home/frame_service/uploaded
RUN mkdir -p /home/frame_service/readconfig
#COPY ./uploaded/Cotier20.jpeg /home/frame_service/uploaded/Cotier20.jpeg
COPY ./readconfig/frame_conf.conf /home/frame_service/readconfig/frame_conf.conf
#COPY ./front  /home/frame_service/front
COPY frameservice /home/frame_service/
WORKDIR /home/frame_service
ENTRYPOINT ["/home/frame_service/frameservice"]
#CMD ["/home/frame_service/frameservice"]