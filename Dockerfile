FROM --platform=${BUILDPLATFORM} alpine AS build-stage

# get target platform
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# get target platform
ARG TARGETOS
ARG TARGETARCH

# install golang
WORKDIR /
RUN GO_VERSION=1.25.0 \
    && wget https://go.dev/dl/go$GO_VERSION.linux-amd64.tar.gz \
    && tar -xzf go$GO_VERSION.linux-amd64.tar.gz \
    && rm go$GO_VERSION.linux-amd64.tar.gz
ENV PATH=$PATH:/go/bin

# set workdir for project
WORKDIR /app
COPY . .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o fireops-edge-printer main.go

# Deploy the application binary into a lean image
FROM ydkn/cups:latest AS build-release-stage

WORKDIR /app

COPY --from=build-stage /app/fireops-edge-printer /app
COPY templates /app/templates

# Install 
RUN sed -i s/deb.debian.org/archive.debian.org/g /etc/apt/sources.list \
    && apt update -y \
    && apt install pandoc -y \
    && apt install chromium -y

# Create startup script
RUN echo '#!/bin/bash\n\
cupsd -f &\n\
sleep 5\n\
/app/fireops-edge-printer\n\
' > /start.sh && chmod +x /start.sh

CMD ["/start.sh"]
