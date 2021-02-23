FROM golang:alpine AS BUILD
ENV TZ=Asia/Ulaanbaatar
ENV GO111MODULE=on
RUN apk add bash ca-certificates git gcc g++ libc-dev
RUN apk add --update tzdata
WORKDIR /go/src/gitlab.com/fibocloud/aws-billing/api_v2
COPY . .
# RUN go get -u github.com/swaggo/swag/cmd/swag
RUN go build

# Stage 2: RUN
FROM alpine
ENV TZ Asia/Ulaanbaatar
RUN apk add --no-cache tzdata ca-certificates 
WORKDIR /home
COPY --from=BUILD /go/src/gitlab.com/fibocloud/aws-billing/api_v2/api /home/
COPY --from=BUILD /go/src/gitlab.com/fibocloud/aws-billing/api_v2/config /home/
# COPY --from=BUILD /go/src/gitlab.com/fibocloud/aws-billing/api/docs /home/
EXPOSE 8081
ENTRYPOINT ["/home/api_v2"]
