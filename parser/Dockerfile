FROM golang:1.25

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /parser-server

EXPOSE 8082

CMD [ "/parser-server" ]