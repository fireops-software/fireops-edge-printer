# FireOps Edge Printer
This repository contains the service, that prints an incomming event on a network printer. Whenever receiving a new event on rabbitmq, the
service converts the incomming event-data into pdf and sends a print job using cups commandline client.

## Configuration
All configuration is done via environmental variables because the intended form of running the project is in a Docker container.

| Variable | Default | Description |
|----------|---------|-------------|
| RABBITMQ_HOST |  | RabbitMQ host (e.g. 192.168.x.x or Hostname) |
| RABBITMQ_PORT | 5672 | RabbitMQ port |
| RABBITMQ_USER |  | RabbitMQ user |
| RABBITMQ_PW |  | RabbitMQ password |
| RABBITMQ_EVENTS_EXCHANGE | fireops-edge-events | RabbitMQ Exchange for events|
| RABBITMQ_EVENTS_ROUTING_KEY | new | RabbitMQ routing key for events |
| RABBITMQ_HEALTH_EXCHANGE | fireops-edge-health | RabbitMQ exchange for health messages |
| RABBITMQ_HEALTH_ROUTING_KEY |  | RabbitMQ routing key for health messages |
||||
| PRINTER_NAME |  | Printer name (name of printer in CUPS) |
| PRINTER_URL |  | Printer ipp url (e.g. ipp://<IP>) |
| PRINTER_DRIVER | everywhere | Printer drivers name (for all ipp printers everywhere) |
| PRINTER_COPIES | 1 | number of copies, that should be printed |
||||
| LOG_LEVEL | INFO | TRACE, DEBUG, INFO, WARNING, ERROR, FATAL, OFF |

### Configuration of CUPS
To be able to manage printers, CUPS is used behind the scence. The Docker image exports port 631 (default CUPS port). To manage printers, this port has to be
forwarded, such that the CUPS UI is available via a web browser.

>**_NOTE:_** Mount a volume for /etc/cups, this will ensure, that printer configuration will be persisted.
