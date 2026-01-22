# QueryLab

QueryLab is a lightweight SQL sandbox built in Go that allows users to safely run and test queries on a PostgreSQL database in isolated sessions. Each user gets a temporary sandboxed environment, making it ideal for learning SQL, testing queries, or providing a safe environment for course databases.


## Features

- Isolated sandbox sessions for each user
- PostgreSQL backend with admin and sandbox roles
- Easy setup with Go and PostgreSQL
- Configurable via `.env` file
- Lightweight and minimal dependencies


## Prerequisites

- Ubuntu 20.04+ (or any Linux/macOS system)
- PostgreSQL installed
- Go installed (version 1.20+ recommended)
- Basic knowledge of SQL and terminal commands


## Installation (Ubuntu)

### 1. Update System Packages

```bash
sudo apt update && sudo apt upgrade -y
```

### 2. Install PostgreSQL

Verify installation:

```bash
psql --version
```

### 3. Install Go

Verify installation:

```bash
go version
```

## Project Setup

1. **Clone the repository**

```bash
git clone git@github.com:pouyatavakoli/QueryLab.git
cd QueryLab
```

2. **Copy and configure environment variables**

```bash
cp .env.example .env
```
edit .env with your actual data


## Database Setup

1. **Create Admin and Sandbox Users**

```bash
sudo -u postgres psql -c "CREATE USER querylab_admin WITH PASSWORD 'admin-strong-password';"
sudo -u postgres psql -c "CREATE USER querylab_sandbox WITH PASSWORD 'sandbox-strong-password';"
```

2. **Create Main Database**

```bash
sudo -u postgres psql -c "CREATE DATABASE querylab OWNER querylab_admin;"
```

3. **Initialize Database Schema**

* Add base tables, sample data, and schema setup to `init.sql`.
* Ensure `INIT_SQL` in `.env` points to this file.


## Running the Server

```bash
go run cmd/server/main.go
```

* The server will start on the port defined in `.env` (default `8080`).
* Open a browser and visit: `http://localhost:8080` to access the query interface.


## Usage

* Enter SQL queries in the web interface.
* Each session runs in a sandboxed environment, isolated from other users.
* Refreshing the page resets your session without affecting the main database or other users.

## Configuration

* `.env` contains all necessary configuration, including database credentials, server port, and initialization file.
* Adjust credentials and paths according to your environment.


## TODO / Future Improvements

* Show only table headers when the table is empty (currently shows nothing)
* Move utility functions out of main files for better code organization


## Security Notes

* Do **not** use weak passwords for database users.
* Consider enabling SSL for PostgreSQL connections in production.
* Sandbox sessions prevent permanent changes to the main database.
