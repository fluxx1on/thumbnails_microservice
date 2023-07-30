FROM golang:1.20

WORKDIR /app

COPY . .
RUN go mod download \
	&& go build -o /bin/server ./cmd

EXPOSE 50051

CMD ["/bin/server"]