# cobra-to-skills

[![Go Report Card](https://goreportcard.com/badge/github.com/yardenshoham/cobra-to-skills)](https://goreportcard.com/report/github.com/yardenshoham/cobra-to-skills)

cobra-to-skills is a utility to generate AI Agent skills from Cobra CLI applications.

# Usage

```bash
cobra-to-skills generate COBRA_BASED_CLI_PATH --output SKILL_DIRECTORY
```

Assuming you have a Cobra CLI application at `./my-cli`, you can generate AI Agent skills with the following command:

```bash
cobra-to-skills generate ./my-cli --output ./skills/my-cli
```

# Build this project

```bash
go build
```

# Run tests

```bash
go test ./...
```

# Run this project

For example you can generate a `kubectl` skill by running

```bash
cobra-to-skills generate kubectl --output ./skills/kubectl
```

## Docker

Docker images are available at
[DockerHub](https://hub.docker.com/r/yardenshoham/cobra-to-skills)
(docker.io/yardenshoham/cobra-to-skills).

Available docker tags

| Tag      | Description                                  |
| -------- | -------------------------------------------- |
| `latest` | latest available release of cobra-to-skills. |
| `va.b.c` | cobra-to-skills version `a.b.c` .            |
| `a.b.c`  | cobra-to-skills version `a.b.c` .            |

### Docker run

```shell script
docker run \
    -v /usr/local/bin:/usr/local/bin:ro \
    -v $PWD:/workdir \
    yardenshoham/cobra-to-skills:latest generate /usr/local/bin/kubectl --output /workdir/skills/kubectl
```

### Docker build

You can build an own docker image by running

```shell
CGO_ENABLED=0 go build && docker build -t cobra-to-skills .
```
