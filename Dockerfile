FROM golang:1.16-alpine

RUN mkdir /app

ADD backend /build/backend
ADD .git/ /build/backend/.git/
WORKDIR /build/backend


RUN go mod vendor

RUN go build -o cocos  -ldflags "-X main.revision=${version} -s -w" ./app

RUN ls -al

CMD ["./cocos"]
