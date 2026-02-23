# Bincang Visual Go

Bincang Visual is a web-based meeting platform built with Flutter and Go. It allows users to create meetings in seconds—no cost, no fees, and no login required. Simply create a new meeting and you’re ready to go!

This project is still under development.
You can try the live demo [here](https://bincang-visual.cloud)

## Setup Instructions

Follow these steps to set up the project::

1. **Clone the repository**: Clone this repository to your local machine
2. **Install Golang**: Ensure that Go is installed on your machine. You can download it from [golang.org](https://go.dev/dl/)
3. **Setup**: Create a .env file in the project root and configure the necessary environment variables.
4. **Install Dependencies**: Navigate to the project directory and install the required Go dependencies:

```
go mod tidy
```

5. **Setup Redis**:

- Start a Redis instance locally or use Docker to run Redis.

6. **Run the Service:** Start the application:

```
go run main.go

```

## Project Structure

```
backend/
├── internal/
│ ├── domain/
│ │ ├── entity/
│ │ ├── repository/
│ │ └── usecase/
│ ├── delivery/
│ │ ├── http/
│ │ └── websocket/
│ ├── repository/
│ │ ├── calendar/
│ │ └── redis/
│ ├── infrastructure/
│ ├── config/
│ └── middleware/
├── pkg/
└── tests/

```
