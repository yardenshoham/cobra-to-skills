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

## Flags

| Flag              | Description                                              |
| ----------------- | -------------------------------------------------------- |
| `--output`, `-o`  | **(required)** Output directory for the generated skill  |
| `--name`          | Override skill name (default: binary name)               |
| `--description`   | Override skill description (default: from root `--help`) |
| `--license`       | License identifier (e.g., `Apache-2.0`)                  |
| `--compatibility` | Compatibility requirements description                   |
| `--allowed-tools` | Allowed tools specification                              |
| `--metadata`      | Metadata key=value pairs (can be repeated)               |
| `--notes`         | Usage notes (can be repeated)                            |

## Example with metadata

```bash
cobra-to-skills generate velero --output ./skills/velero \
  --name velero \
  --description "Back up and restore Kubernetes cluster resources" \
  --license Apache-2.0 \
  --compatibility "Requires a running Kubernetes cluster with kubectl configured" \
  --allowed-tools "Bash(velero:*) Bash(kubectl:*) Read" \
  --metadata author=vmware-tanzu \
  --metadata category=kubernetes-backup \
  --notes 'Use `--wait` on `backup create` to block until the operation completes.'
```

## Example output

Running `cobra-to-skills generate ./cobra-to-skills --output ./skills` produces:

```
skills/
├── SKILL.md
└── references/
    ├── cobra-to-skills.md
    ├── cobra-to-skills_completion.md
    ├── cobra-to-skills_generate.md
    ├── cobra-to-skills_version.md
    └── ...
```

**SKILL.md:**

```markdown
---
name: cobra-to-skills
description: "Generate AI Agent skills from Cobra CLI applications"
---

# cobra-to-skills

Generate AI Agent skills from Cobra CLI applications

## Available Commands

- [`cobra-to-skills generate`](references/cobra-to-skills_generate.md) - Generate AI Agent skills from a Cobra CLI binary
- [`cobra-to-skills version`](references/cobra-to-skills_version.md) - Print the version of cobra-to-skills

See [references/cobra-to-skills.md](references/cobra-to-skills.md) for root command flags.

Run `cobra-to-skills --help` or `cobra-to-skills <command> --help` for full usage details.
```

**references/cobra-to-skills_version.md:**

````markdown
# cobra-to-skills version

Print the version of cobra-to-skills

```
cobra-to-skills version [flags]
```

## Examples

```
cobra-to-skills version
```

### Options

```
  -h, --help   help for version
```
````

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
