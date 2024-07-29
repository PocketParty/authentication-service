# authentication-service

Golang authentication server

## Setup PostgreSQL Server

This server requires a postgresql server to be running.

After setting up the database, run the following

```bash
export DB_USER=<db-username>
export DB_NAME=<database name>
export DB_PASSWORD=<password>
export DB_HOST=<address of DB> # default is localhost if you are running locally
export DB_PORT=<port> # default is 5432
export DB_SSL_MODE=<ssl-mode> # default is disable
```

## Running the server

In a machine with Go installed:

```bash
go run main.go
```

### Alternative: Running the server in a docker container

(Optional) Build the docker image

```bash
docker build -t authentication-service .
```

Or pull the image from github registry

```bash
docker pull ghcr.io/pocketparty/authentication-service:latest
```

Run the docker container

```bash
docker run -p 8080:8080 --name authentication-server-demo -e DB_USER=<db-username> -e DB_NAME=<database name> -e DB_PASSWORD=<password> -e DB_HOST=<address of DB> -e DB_PORT=<port> -e DB_SSL_MODE=<ssl-mode> ghcr.io/pocketparty/authentication-service
```

add the flag --network="host" if you want to run the container in the host network

## Communication with the server

The server has a REST API, it can create new users, and authenticate existing users.

The following examples use curl to show how to test the requests.

### Create a new user

The endpoint to create a new user is /signup

It expects a POST request with a JSON body containing the username and password of the new user.

```bash
 curl -X POST http://localhost:8080/signup -d '{"username":"test-username", "password":"test-pass"}' -H "Content-Type: application/json"
```

Replace localhost with the address of the server if it is running on a different machine.

If the request is successful, the server will return a 200 status code.

### Authenticate a user

The endpoint to authenticate a user is /signin

It expects a GET request where the username and password are passed as query parameters.

It returns a JWT token if the user is authenticated.

```bash
curl -X GET http://localhost:8080/signin -H "Content-Type: application/json" -d '{"username":"test-username", "password":"test-pass"}'
```
