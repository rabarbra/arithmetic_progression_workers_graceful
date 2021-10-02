## Arithmetic progression workers
### Requirements
* go 1.17
* [go-playground/validator/v10](https://github.com/go-playground/validator) v10.9.0
* [gorilla/mux](https://github.com/gorilla/mux) v1.8.0
* [stretchr/testify](https://github.com/stretchr/testify) v1.7.0
### Installation
```bash
$ git clone https://github.com/rabarbra/arithmetic_progression_workers_test.git
$ cd arithmetic_progression_workers_test
$ go build
```
### Running server
```bash
$ # Run server with n max parallel workers allowed (by default n = 2)
$ ./workers_server -n 5
```
### Usage
Add new task, where:
* n - number of elements in progression (positive int)
* d - delta between adjacent elements (float)
* n1 - first element (float)
* I - time interval between iterations (seconds, positive float)
* TTL - time serving done tasks before deleting (seconds, positive float)
```bash
$ curl -X POST 127.0.0.0:8000/add -d'{"n":30,"d":1,"n1":0,"I":2,"TTL":50}'
```

Get all tasks sorted by status:
```bash
$ curl -X GET 127.0.0.0:8000/get
```