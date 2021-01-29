FROM ubuntu:18.04
WORKDIR /app
ADD build/tenkai-api /app
ADD app.yaml /app
CMD ["/app/tenkai-api"]
