 FROM alpine:3.4

 ADD conti /conti

 ENTRYPOINT ["/conti"]
