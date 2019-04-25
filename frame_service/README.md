# Frame_service api

## Getting Started
This is service for frame devices. It run in container as daemon
## To build service
```
make build
```
docker build -t framedoc .


## For run docker
docker run -dit --name framer2 -p 8080:8080
