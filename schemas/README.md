# Ecommerce System Schema
This repository contains the SQL table creation schema for the ecom database.

## Postgres
The schema for Postgresql is located in the `postgres` directory.

### Creating the database
Ensure you have set the environment variables for `PGHOST`, `PGUSER`, `PGPASSWORD`, `PGPORT`.

The shell scripts are hard-coded to create the schemas inside a datagbase called `ecom_dev`. If you are creating tables for production you will need to adjust the database name appropriately. This is to prevent accidental deletion of production schemas.

1. First create a new database called `ecom_dev` in Postgres:

    ```sql
       CREATE DATABASE ecom_dev;
    ```

2. Run the `create-postgres-schema.sh` shell script to create the database tables for the ecom system.

    ```sh
    $ ./scripts/create-postgres-schema
    ```

### Deleting the database

1. Run the `drop-postgres-schema.sh` shell script to delete the tables for the ecom system.

    ```sh
    $ ./scripts/drop-postgres-schema
    ```
