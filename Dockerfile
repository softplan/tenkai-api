FROM ubuntu:18.04
WORKDIR /app
ADD build/tenkai-api /app
CMD ["/app/tenkai-api"]
