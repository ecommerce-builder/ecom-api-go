FROM alpine:latest
RUN apk add ca-certificates

# Make an app directory for our API app
RUN mkdir /app

# Copy Ecom API app to the app directory
COPY ./bin/ecom-api-alpine-amd64 /app

# Make the executable runnable
RUN chmod 0744 /app/ecom-api-alpine-amd64

CMD [ "/app/ecom-api-alpine-amd64" ]
