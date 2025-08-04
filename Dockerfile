FROM golang:alpine3.22

RUN adduser -D -s /bin/ash -u 1000 devuser

RUN apk update && apk --no-cache add autoconf \
    curl \
    git

# Air for hot-reloading the example
RUN go install github.com/air-verse/air@latest

# Install Go tools
RUN go install github.com/cweill/gotests/gotests@latest
RUN go install github.com/fatih/gomodifytags@latest
RUN go install github.com/haya14busa/goplay/cmd/goplay@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install honnef.co/go/tools/cmd/staticcheck@latest
RUN go install golang.org/x/tools/gopls@latest

# Install CLI tools
RUN go install golang.org/x/tools/cmd/stringer@latest
RUN go install github.com/spf13/cobra-cli@latest

RUN chown -Rf devuser /go
USER devuser