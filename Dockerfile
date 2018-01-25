FROM scratch
ADD bin/cts /usr/local/bin/
ADD ca-certificates.crt /etc/ssl/certs/
CMD ["cts"]
