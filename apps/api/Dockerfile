FROM golang:1.24-alpine AS build

ENV GOTOOLCHAIN=auto

WORKDIR /repo/apps/api

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api/ .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /api .

FROM gcr.io/distroless/static-debian12

COPY --from=build /api /api

EXPOSE 4000

ENTRYPOINT ["/api"]
