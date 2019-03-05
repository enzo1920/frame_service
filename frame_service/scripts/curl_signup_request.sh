#!/bin/sh
curl -X POST  -H "Accept: Application/json" -H "Content-Type: application/json" http://188.227.18.141:8080/v1/signup -d '{"username": "rickrou","password": "mysecurepassword"}'
