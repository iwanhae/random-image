FROM golang:1.17-buster as server
WORKDIR /app/random-image
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o /random-image .
RUN chmod +x /random-image

FROM node:16 as web
COPY web/. /app/web/.
WORKDIR /app/web
RUN yarn && yarn generate


FROM ubuntu:18.04
COPY --from=server /random-image /random-image
COPY --from=web /app/web/out/. /web/out/.
EXPOSE 8080
ENTRYPOINT [ "/random-image" ]