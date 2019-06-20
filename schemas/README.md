# Ecommerce System Schema
This repository contains the SQL table creation schema for the ecom database.


## Postgres
The schema for Postgresql is located in the `postgres` directory.


### Creating the database

Ensure you have set the environment variables for `PGHOST`, `PGUSER`, `PGPASSWORD`, `PGPORT`.

The shell scripts are hard-coded to create the schemas inside a datagbase called `ecom_dev`. If you are creating tables for production you will need to adjust the database name appropriately. This is to prevent accidental deletion of production schemas.

1. First create a new Postgres database named `ecom_dev`:

``` sql
$ psql
CREATE DATABASE ecom_dev;
```


2. Run the `create-postgres-schema.sh` shell script to create the database tables for the ecom system.

``` bash
$ ./scripts/create-postgres-schema
```


3. (Optional Step) Load the demo data to populate the database.

``` bash
$ ./scripts/load-demo-data
```


### Deleting the database

1. Run the `drop-postgres-schema.sh` shell script to delete the tables for the ecom system.

``` bash
$ ./scripts/drop-postgres-schema
```
