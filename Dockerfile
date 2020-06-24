FROM golang:alpine
LABEL maintainer="joseph@proton.ai"
RUN mkdir /app 
ADD . /app/
WORKDIR /app
RUN go build -o main .

EXPOSE 8100

CMD ["./main"]