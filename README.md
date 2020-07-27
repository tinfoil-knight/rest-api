## Directory Structure
```
.
|____contact.go
|____contact__test.go
|____config
| |____config.go
|____models
| |____models.go
|____helpers
| |____cache.go
| |____mongodb.go
|____go.sum
|____go.mod
|____config.yaml
|____README.md
|____.air.conf
|____.gitignore
```

## Application Structure
TBD

## Installation
Fork the repo and pull it to your workstation.

A `config.yaml` file is needed to run the application. You'll need to install `Status Bot` for updates in slack. Click the button below to add the `Status Bot` to your Slack workspace.

<a href="https://slack.com/oauth/v2/authorize?client_id=1263601250499.1262025136661&scope=incoming-webhook&user_scope="><img alt="Add to Slack" height="40" width="139" src="https://platform.slack-edge.com/img/add_to_slack.png" srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" /></a>


Example YAML configuration
```yaml
---
COLLECTION: contacts
DB: phonebook
TESTDB: phonebook-test
MONGODB-URI: "mongodb://localhost:27017"
PORT: "8080"
SLACK-HOOK: <Your Slack-Webhook-URL goes here>
```

## Running Tests
```bash
go test -v
```
Current Test Coverage: 51.4%

### Author
Kunal Kundu [@tinfoil-knight](https://github.com/tinfoil-knight)

### Acknowledgements
Matthias [@qarchmage](https://github.com/qarchmage) for fixing race condition in database while testing.

### License
TBD