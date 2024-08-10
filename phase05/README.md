
# Tracegoute

This project is a Go-based web application designed to perform network traceroute operations via a RESTful API. It utilizes the Gin framework for handling HTTP requests and offers flexibility in deployment through environment variable configuration.

# Features
- Traceroute functionality exposed via an HTTP API.
- Configurable server settings using environment variables.
- Lightweight and fast, built with the Gin web framework.

# Installation

## Prerequisites

- Go 1.16 or higher
- Git

## Clone the Repository

```bash
git clone https://github.com/Star-Academy/Summer1403-Devops-Team11.git
cd phase05
```

## Install Dependencies
Ensure you have Go modules enabled, and then run:

```bash
go mod tidy
```

## Environment Setup
Modify the .env file to set your desired configuration, such as the server host.

## Running the Application
Start the server using the following command:

```bash
go run main.go
```
The server will start on the host and port defined in the .env file, defaulting to :8080 if not specified.

## Usage

### API Reference

```http
  GET /traceroute/{domain}
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `domain`      | `string` | **Required**. domain for traceroute |

Example using curl:

```bash
curl http://localhost:8080/traceroute/example.com
```

## Project Structure
- main.go: Entry point of the application.
- handler: Contains the HTTP handler functions.
- helper: Utility functions and helpers for environment management and other tasks.

## Acknowledgments
- [Gin](https://gin-gonic.com/) Web Framework
- [Godotenv](https://github.com/joho/godotenv)
