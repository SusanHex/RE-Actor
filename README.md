# RE-Actor

Hi! Thanks for coming to look at RE-Actor! This application is intended to apply a Regular Expression based pattern to the logs of a Docker container. If the pattern matches the log, then the values are copied into the supplied template. Once complete, the resulting value will be sent to the selected action. Currently, you can post a message to a Discord webhook or send an email through SMTP.

*Example Docker Compose*
```
services:
  reactor:
    image: susanhex/reactor:latest
    container_name: reactor
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DISCORD_WEBHOOK_URLS=<insert your Discord Webhook URL here>
      - LOG_LEVEL=${LOG_LEVEL:-INFO}
      - CONTAINER_NAME=vintage-story-server
      # This pattern will match when players leave/disconnect/join the vintage story server.
      - PATTERN=(:?\[Server\sEvent]\s(?:Player\s)?(?P<Player>\S+)\s(?:\[\S+\s)?(?P<Verb>(?:joins)|(?:left)))|(?:\[Server\sNotification]\sUDP\:\sclient\s(?<Verb>\S+)\s(?<Player>\S+)) 
      # This template will create messages like this: "SusanHex left"
      - TEMPLATE=$${Player} $${Verb}
```
In order to run this, place the above text in a file named `compose.yml` and type `docker compose up` within the same folder.

##Configuration
Here is a list of environme*nt variables that are used to configure RE-Actor:

**Required**
| Variable Name | Description |
|---------------|-------------|
| `CONTAINER_NAME` | This is the name of the container you would like RE-Actor to monitor. It should only match one container. |
| `PATTERN` | This is the Regular Expression pattern to run against each log from the respective container. Please ensure that this pattern is valid and does what you expect. Here is a great resource for creating your pattern: [Regex101](https://regex101.com). |
| `TEMPLATE` | if the current log message matches your pattern, this defines how the output will look. Please refer to the previous linked resource for help on this. |

**Optional**
| Variable Name | Description |
|---------------|-------------|
| `ACTION_NAME` | The name of the action you are using, will default to `discord_webhook` if not provided. Valid options are: `discord_webhook`, `smtp`, or `test`. |
| `LOG_LEVEL` | Determines how verobose the logging will be. Here are the valid options: `DEBUG`, `INFO`, `WARNING`, and `ERROR`. The default value is `INFO`, if not set. |

**Discord Webhook**
| Variable Name | Description |
|---------------|-------------|
| `DISCORD_WEBHOOK_URLS` | These are the webhooks that RE-Actor will post the resulting message to. Mutliple URLs can be specified like this: `<URL 1><URL Separator><URL 2>`. With the default configuration, a multi-URL example would be`https:...;;;https:...`.|
| `DISCORD_WEBHOOK_URL_SEPARATOR` | This character set separates the URLs in the `DISCORD_WEBHOOK_URLS` variable. If not set, the default is `;;;`. |

**SMTP**
| Variable Name | Description |
|---------------|-------------|
| `SMPTP_HOST` | The SMTP host you are using. |
| `SMTP_PORT` | SMTP port you wish to use. |
| `SMTP_SEND_FROM` | Email that will be sending the messages. |
| `SMTP_SEND_TO` | Email that will receive the messages. |
| `SMTP_PASSWORD` | Password for the SEND_FROM email. |
| `SMTP_SUBJECT` | Subject for the emails being sent out. |

**Test**
| Variable Name | Description |
|---------------|-------------|
| `TEST_ACTION_DELAY` | This is the delay before logging the test action message in miliseconds. Default is `0`. |
