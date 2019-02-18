# PowerShell
# docker-compose build; docker-compose up -d
FROM golang:1.10

WORKDIR /go/src/app
COPY . .

#EXPOSE 8081

RUN go get -d ./...
RUN go install ./...

#RUN go run send.go
CMD go run worker/worker.go worker/db.go


