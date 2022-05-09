FROM golang:1.18-alpine

# Install LaTeX and build dvisvgm from source
RUN apk add --no-cache --virtual .build-deps \
    build-base autoconf automake libtool texlive-dev freetype-dev brotli-dev woff2-dev && \
    cd /root && \
    wget https://github.com/mgieseki/dvisvgm/archive/2.13.3.tar.gz && \
    tar zxvf 2.13.3.tar.gz && \
    cd dvisvgm-2.13.3 && \
    ./autogen.sh && ./configure && make && make install && \
    cd ../ && rm -fr dvisvgm* *.tar.gz && \
    apk del .build-deps
RUN apk add texlive-full woff2 ghostscript inkscape

# Set up directory
RUN mkdir /var/app
ADD ./src /var/app
WORKDIR /var/app

# Install dependencies
RUN go mod download

# Build with libraries
RUN go build -o texbot .

# Define entrypoint to texbot
ENTRYPOINT ["/var/app/texbot"]
