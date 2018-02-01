FROM centurylink/ca-certs
ADD bin/cts /usr/local/bin/
CMD ["cts"]
