-- This file is executed by Postgres on first startup

-- Create admin role
CREATE ROLE querylab_admin LOGIN;
ALTER ROLE querylab_admin CREATEDB;

-- Create sandbox role
CREATE ROLE querylab_sandbox LOGIN;
ALTER ROLE querylab_sandbox
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOINHERIT
    NOREPLICATION
    CONNECTION LIMIT 50;
