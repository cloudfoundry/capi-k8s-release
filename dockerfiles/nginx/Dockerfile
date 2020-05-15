FROM ubuntu:bionic

RUN mkdir -p /nginx
WORKDIR /nginx
COPY *.tar.gz ./
COPY *.patch ./

# install only what is needed to run everything but `./configure` for build optimization
RUN apt-get update && apt-get install -y \
  patch

RUN tar xzvf nginx-upload-module-2.3.0.tar.gz
RUN cd nginx-upload-module-2.3.0 && \
  patch < ../upload_module_put_support.patch

RUN tar xzvf nginx-1.16.1.tar.gz
RUN tar xzvf pcre-8.42.tar.gz

# install whatever is needed for "./configure"
RUN apt-get install -y \
  gcc \
  g++ \
  libssl-dev \
  make \
  openssl \
  zlibc \
  zlib1g \
  zlib1g-dev

RUN cd nginx-1.16.1 && \
  ./configure \
    --with-debug \
    --with-pcre=../pcre-8.42 \
    --add-module=../nginx-upload-module-2.3.0 && \
  make -j $(nproc) && \
  make -j $(nproc) install

CMD ["/usr/local/nginx/sbin/nginx", "-g", "daemon off;", "-c", "/etc/nginx/nginx.conf" ]
