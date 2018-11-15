FROM alpine:latest
RUN apk add ca-certificates
RUN mkdir /app

# Google service account credentials
COPY ./ecom-test-fa3e406ce4fe.json /app

# Postgres SSL certificate files
COPY ./certs/pg/ /app/certs/pg

# Postgres SSL certificates (from GCP control panel)
ENV PGSSLCERT /app/certs/pg/client-cert.pem
ENV PGSSLROOTCERT /app/certs/pg/server-ca.pem
ENV PGSSLKEY /app/certs/pg/client-key.pem
ENV PGSSLMODE verify-ca

# Ecom API
COPY ./ecom-api-go /app

# Make the executable runnable
RUN chmod 0744 /app/ecom-api-go

CMD [ "/app/ecom-api-go" ]
