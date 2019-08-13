FROM ubuntu:18.04
#ADD ca-certificates.crt /etc/ssl/certs/
WORKDIR app
#ADD .helm/ /app/.helm/
ADD build/tenkai-api /app
ADD app.yaml /app
CMD ["/app/tenkai-api"]
