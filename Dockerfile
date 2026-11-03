FROM golang AS runner

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go mod download

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV PORT=3000
EXPOSE 3000

RUN go build -o app .
ENTRYPOINT ["./app"]