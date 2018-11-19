FROM alpine:latest
RUN apk add ca-certificates

# Make an app directory for our API app
RUN mkdir /app

# Copy Postgres SSL certificate files (From Google Cloud Console)
COPY ./certs/pg/ /app/certs/pg

# Copy Ecom API app to the app directory
COPY ./ecom-api-go-alpine /app

# Make the executable runnable
RUN chmod 0744 /app/ecom-api-go-alpine

CMD [ "/app/ecom-api-go-alpine" ]
