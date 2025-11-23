# RE-Actor

Hi! Thanks for coming to look at RE-Actor! This application is intended to apply a Regular Expression based pattern to the logs of a Docker container. If the pattern matches the log, then the values are copied into the supplied template. Once complete, the resulting value will be sent to the selected action. Currently, you can post a message to a Discord webhook or send an email through SMTP.

Here is a list of environment variables that are used to configure RE-Actor:

**Required**
| Variable Name | Description |
|---------------|-------------|
| `CONTAINER_NAME` | This is the name of the container you would like RE-Actor to monitor. It should only match one container. |
| `PATTERN` | This is the Regular Expression pattern to run against each log from the respective container. Please ensure that this pattern is valid and does what you expect. Here is a great resource for creating your pattern: [Regex101](https://regex101.com). |
| `TEMPLATE` | if the current log message matches your pattern, this defines how the output will look. Please refer to the previous linked resource for help on this. |

**Optional**
| Variable Name | Description |
|---------------|-------------|
| `ACTION_NAME` | The name of the action you are using, will default to discord_webhook if not provided. Valid options are: `discord_webhook` or `smtp`. |
| `LOG_LEVEL` | Determines how verobose the logging will be. Here are the valid options: `DEBUG`, `INFO`, `WARNING`, and `ERROR`. The default value is `INFO`, if not set. |

**Discord Webhook**
| Variable Name | Description |
|---------------|-------------|
| `DISCORD_WEBHOOK_URL` | This is the webhook that RE-Actor will post the resulting message to. |

**SMTP**
| Variable Name | Description |
|---------------|-------------|
| `SMPTP_HOST` | The SMTP host you are using. |
| `SMTP_PORT` | SMTP port you wish to use. |
| `SEND_FROM` | Email that will be sending the messages. |
| `SEND_TO` | Email that will receive the messages. |
| `PASSWORD` | Password for the SEND_FROM email. |
| `SUBJECT` | Subject for the emails being sent out. |
