FROM golang:latest 
LABEL MAINTAINER='Venkatesh Laguduva <lbvenkatesh@gmail.com>'
RUN mkdir /app 
ADD . /app/
RUN apt -y update && apt -y install git
RUN go get -u github.com/gorilla/mux
RUN go get -u github.com/lib/pq
WORKDIR /app 
RUN go build -o main . 
ENV PORT=8080
ENV GIN_MODE=release
EXPOSE 8080
CMD ["/app/main"]