FROM ubuntu:bionic AS nginx-compiler

RUN apt-get update && apt-get install -y \
  patch \
  gcc \
  g++ \
  libssl-dev \
  make \
  openssl \
  zlibc \
  zlib1g \
  zlib1g-dev \
  wget

WORKDIR /capi-release-dir
COPY build.sh .
RUN ./build.sh

FROM ubuntu:bionic

RUN apt-get update && apt-get install -y \
  openssl

COPY --from=nginx-compiler /usr/local/nginx /usr/local/nginx

RUN groupadd -g 1000 cnb
RUN useradd -r -u 1000 -g 1000 cnb

ENTRYPOINT [ "/usr/local/nginx/sbin/nginx" ]
