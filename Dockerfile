FROM node:lts-alpine AS ui_builder
RUN apk update && apk add git
ARG VER
WORKDIR /dashboard
ADD dashboard .
RUN npm i; npm run build; 


FROM golang:1.23 AS go_builder
ARG VER
WORKDIR /server
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build


FROM scratch
LABEL org.opencontainers.image.source="https://github.com/robrotheram/warptail"
COPY --from=ui_builder /dashboard/dist /dashboard/dist
COPY --from=go_builder /server/warptail /go/bin/warptail
ENTRYPOINT ["/go/bin/warptail"]
