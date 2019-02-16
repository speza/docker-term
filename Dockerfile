FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o term .
EXPOSE 8080
CMD ["/app/term"]