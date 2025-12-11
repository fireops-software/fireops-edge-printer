FROM ydkn/cups:latest

# install golang
WORKDIR /
RUN GO_VERSION=1.25.5 \
    && wget https://go.dev/dl/go$GO_VERSION.linux-amd64.tar.gz \
    && tar -xzf go$GO_VERSION.linux-amd64.tar.gz \
    && rm go$GO_VERSION.linux-amd64.tar.gz

ENV PATH=$PATH:/go/bin
RUN go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install golang.org/x/tools/gopls@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest

# Install 
RUN sed -i s/deb.debian.org/archive.debian.org/g /etc/apt/sources.list \
    && apt update -y \
    && apt install pandoc -y \
    && apt install chromium -y

CMD ["cupsd -f &"]
