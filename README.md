# Cloud File System

This is a simple project that expose file system operations through a rest api

## Run it

### Requirements
Have go installed, [here's how](https://go.dev/doc/install).

### Instructions
The server is configured with 2 environment variables :
- STORAGE_DIR: format -> /path/to/dir | example -> /home/jesus/Documents/ , defines the exposed path
- PORT: format -> :PORT | example -> :8080, port on which the server will be listening, default -> :8080

Set STORAGE_DIR is required to run the server.

If you use bash you can do the following:
``` bash
export STORAGE_DIR=/path/to/dir
go build -o main ./src/cmd/server/main.go && ./main
```

## Usage
The server expose the route /browse/ as the root which is the STORAGE_DIR variable, for example /browse/memes/ map to STORAGE_DIR/memes in the file system
