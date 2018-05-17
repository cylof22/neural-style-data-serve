FROM ubuntu:16.04

WORKDIR app

ADD ../../bin/neural-style-image-store /app 

ENV PATH $PATH:/app

CMD ["neural-style-image-store"]

EXPOSE 5000