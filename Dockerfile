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

# ---------------------------------------------------------------------------------------
# Deploy the application binary into a lean image
# ---------------------------------------------------------------------------------------
FROM debian:bookworm-slim AS build-release-stage

WORKDIR /app

COPY --from=build-stage /app/fireops-edge-printer /app
COPY templates /app/templates

# Install 
RUN apt update -y \
    && apt install wget -y \
    && apt install pandoc -y \
    && apt install chromium -y \
    && apt install cups -y \
    && apt install printer-driver-all -y \
    && apt install sed -y

# Create CUPS groups if they don't exist and add user
RUN groupadd -f lp && \
    groupadd -f lpadmin && \
    # Create user with home directory
    useradd -m -u 1001 -g lp -G lpadmin -s /bin/bash fireops && \
    # Set password
    echo "fireops:fireops" | chpasswd


# Configure CUPS to listen on all interfaces
RUN sed -i 's/Listen localhost:631/Listen 0.0.0.0:631/' /etc/cups/cupsd.conf

# Add ServerAlias and WebInterface
RUN sed -i '/Listen 0.0.0.0:631/a ServerAlias *' /etc/cups/cupsd.conf && \
    sed -i '/ServerAlias \*/a WebInterface Yes' /etc/cups/cupsd.conf

# Replace the entire Location blocks with permissive settings
RUN sed -i '/<Location \/>/,/<\/Location>/c\
<Location />\n\
  Order allow,deny\n\
  Allow all\n\
</Location>' /etc/cups/cupsd.conf

RUN sed -i '/<Location \/admin>/,/<\/Location>/c\
<Location /admin>\n\
  Order allow,deny\n\
  Allow all\n\
</Location>' /etc/cups/cupsd.conf

RUN sed -i '/<Location \/admin\/conf>/,/<\/Location>/c\
<Location /admin/conf>\n\
  Order allow,deny\n\
  Allow all\n\
</Location>' /etc/cups/cupsd.conf

# Create spool directory and set permissions
RUN mkdir -p /var/spool/cups && \
    chown -R root:lp /etc/cups /var/spool/cups && \
    chmod -R 755 /etc/cups /var/spool/cups

# Expose CUPS web interface port
EXPOSE 631/tcp

# Create startup script
RUN echo '#!/bin/bash\n\
cupsd -f &\n\
sleep 5\n\
/app/fireops-edge-printer\n\
' > /start.sh && chmod +x /start.sh

CMD ["/start.sh"]
