# Authentication Service

Simple authentication service with support for JWT tokens.

## Overview

Auth Service is a Golang application that provides a simple API for user authentication. 
It uses JWT tokens to authenticate users and protect endpoints.


## Table of Contents

- [Usage](#usage)
- [Configuration](#configuration)
- [Endpoints](#endpoints)

## Usage

The easiest way to get the service up and running is to use Docker. Run the following command from the root of the project:

`docker-compose up`

## Configuration

The following environment variables can be used to configure the application:

| Name | Description | Default |
| --- | --- | --- |
| `DB_HOST` | Hostname for the database server | `localhost` |
| `DB_PORT` | Port for the database server | `26257` |
| `DB_NAME` | Name of the database | `auth` |
| `DB_USER` | Username for the database | `root` |
| `DB_PASSWORD` | Password for the database |  |
| `DB_SSL_MODE` | SSL mode for the database connection | `disable` |
| `JWT_SECRET_KEY` | Secret key for JWT tokens | `123` |
| `TOKEN_TTL` | Time to live for JWT tokens | `30m` |
| `PORT` | Port for the HTTP server | `8080` |
| `MIGRATION_PATH` | Path to database migrations | `/app/migrations` |

## Endpoints

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/register` | Register a new user |
| `POST` | `/login` | Login to the application |
| `POST` | `/refresh` | Refresh a JWT token |
| `POST` | `/validate` | Validate a JWT token |
| `POST` | `/change-password` | Change the password for a user |

`/register` - User registration

```
curl --request POST \
  --url http://localhost:8080/register \
  --header 'Content-Type: application/json' \
  --data '{
	"username": "newuser",
	"password": "newpassword"
}'
```

`/login` - User login

Request:

```
curl --request POST \
  --url http://localhost:8080/login \
  --header 'Content-Type: application/json' \
  --data '{
	"username": "newuser",
	"password": "newpassword"
}'
```

`/refresh` - Refresh access token

Request:

```
curl --request POST \
  --url http://localhost:8080/refresh \
  --header 'Content-Type: application/json' \
  --data '{
	"refresh_token": "<token>"
}'
```

`/validate` - Validate JWT

Request:

```
curl --request POST \
  --url http://localhost:8080/validate \
  --header 'Content-Type: application/json' \
  --data '{
	"token": "<token>"
}'
```

`/change-password` - Change password

Request:

```
curl --request POST \
  --url http://localhost:8080/change-password \
  --header 'Content-Type: application/json' \
  --header 'Authorization: Bearer <token>' \
  --data '{
	"old_password": "newpassword",
	"new_password": "newpassword2"
}'
```
