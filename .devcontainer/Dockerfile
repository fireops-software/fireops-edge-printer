FROM debian:bookworm-slim

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

# Add ServerAlias directive
RUN echo "ServerAlias *" >> /etc/cups/cupsd.conf

# Configure access controls using correct syntax
# Use Allow/Deny directives instead of Require all granted
RUN sed -i '/<Location \/>/,/<\/Location>/ s/AuthType Basic/AuthType None/' /etc/cups/cupsd.conf && \
    sed -i '/<Location \/>/,/<\/Location>/ s/Require user @SYSTEM/Allow @OWNER/' /etc/cups/cupsd.conf

RUN sed -i '/<Location \/admin>/,/<\/Location>/ s/AuthType Basic/AuthType None/' /etc/cups/cupsd.conf && \
    sed -i '/<Location \/admin>/,/<\/Location>/ s/Require user @SYSTEM/Allow @OWNER/' /etc/cups/cupsd.conf

RUN sed -i '/<Location \/admin\/conf>/,/<\/Location>/ s/AuthType Basic/AuthType None/' /etc/cups/cupsd.conf && \
    sed -i '/<Location \/admin\/conf>/,/<\/Location>/ s/Require user @SYSTEM/Allow @OWNER/' /etc/cups/cupsd.conf

# Create spool directory and set permissions
RUN mkdir -p /var/spool/cups && \
    chown -R root:lp /etc/cups /var/spool/cups && \
    chmod -R 755 /etc/cups /var/spool/cups

# Expose CUPS web interface port
EXPOSE 631/tcp

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


