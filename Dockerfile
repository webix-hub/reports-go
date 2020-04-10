FROM centurylink/ca-certs
WORKDIR /app
COPY ./metadb /app

CMD ["/app/metadb"]