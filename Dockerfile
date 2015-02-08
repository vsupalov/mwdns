FROM golang:1.3-onbuild

RUN apt-get update -qq && \
    apt-get install -qqy npm && \
    npm install -g less uglify-js && \
    rm -rf /var/lib/apt/lists/*

RUN make

CMD ["app", "--addr", "0.0.0.0:8000"]
