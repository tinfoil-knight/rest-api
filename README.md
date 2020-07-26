`.gitignore` the `config.yaml` file when using sensitive resources.

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

## Structure

## Installation
```bash
go get github.com/tinfoil-knight/rest-api
```

## Running Tests
```bash
go test -v
```

### Author
Kunal Kundu [@tinfoil-knight](https://github.com/tinfoil-knight)

### Acknowledgements
Matthias [@qarchmage](https://github.com/qarchmage) for fixing concurrency issues faced while using the database for tests.

### License
