* Demo for mix of HTTP and gRPC in Go
* Warning: I had to use apache thrift 0.10.0 (made it with 'git checkout tags/0.10.0' in $GOPATH/src).
* For jaeger,
  * docker run -d -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 -p5775:5775/udp -p6831:6831/udp -p6832:6832/udp \
  -p5778:5778 -p16686:16686 -p14268:14268 -p9411:9411 jaegertracing/all-in-one:latest
    * https://jaeger.readthedocs.io/en/latest/getting_started/#all-in-one-docker-image
