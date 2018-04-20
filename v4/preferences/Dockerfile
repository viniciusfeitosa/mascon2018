FROM golang:latest 

LABEL Name=preferences_service Version=0.0.1 

RUN  mkdir -p /go/src \
  && mkdir -p /go/bin \
  && mkdir -p /go/pkg
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH 

# now copy your app to the proper build path
RUN mkdir -p $GOPATH/src/app 
ADD . $GOPATH/src/app

WORKDIR $GOPATH/src/app
RUN go build -o main . 
CMD ["/go/src/app/main"]

EXPOSE 5000 50051