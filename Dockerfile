 FROM alpine:3.4

 ADD contd /contd

 ENTRYPOINT ["/contd"]
