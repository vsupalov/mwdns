FROM golang:1.3-onbuild

RUN apt-get update && \
    apt-get install npm && \
    npm install -g less uglify-js && \
    rm -rf /var/lib/apt/lists/*

RUN make

CMD ["app", "0.0.0.0:8000"]
