-- Run this once as postgres
-- Admin role (trusted)
CREATE ROLE querylab_admin LOGIN PASSWORD 'admin-strong-password';
ALTER ROLE querylab_admin CREATEDB;

-- Sandbox role (untrusted)
CREATE ROLE querylab_sandbox LOGIN PASSWORD 'sandbox-strong-password';
ALTER ROLE querylab_sandbox
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOINHERIT
    NOREPLICATION
    CONNECTION LIMIT 50;
