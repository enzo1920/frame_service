FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install -y golang


ENV PORT 8080
EXPOSE $PORT

COPY frameservice /
CMD ["/frameservice"]