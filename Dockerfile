# build
FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./

COPY . .

RUN go build -o main

# run
FROM chromedp/headless-shell:latest

RUN apt update -y
RUN apt install dumb-init

ENTRYPOINT ["dumb-init", "--"]

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/assr.yaml .

EXPOSE 3000

ENV PATH=$PATH:/headless-shell
CMD ["./main", "run", "-c=assr.yaml"]