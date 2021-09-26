FROM golang:1.17-buster as server
RUN apt update && apt install -y libwebp-dev
WORKDIR /app/random-image
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o /random-image .
RUN chmod +x /random-image

FROM node:16 as web
COPY web/package.json /app/web/package.json
WORKDIR /app/web
RUN yarn
COPY web/. /app/web/.
RUN yarn generate


FROM golang:1.17-buster
RUN apt update && apt install -y libwebp-dev
WORKDIR /
COPY --from=server /random-image /random-image
COPY --from=web /app/web/out/. /web/out/.
EXPOSE 8080
ENTRYPOINT [ "/random-image" ]