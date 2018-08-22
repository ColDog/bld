# Configuration

View the [examples](../examples/) directory for a list of examples.

## Types

- `Source`: A set of files or a directory as an input to a build.
- `Step`: A docker image and a set of instructions to run.
- `Volume`: A directory to mount inside of a container, will not be inspected.
- `Import`: A source to be mounted in a step. If the source is changed, the step
  will be rebuilt.
- `Export`: A source created from work done inside a container.

## Configuration File

Builds are configured by using a `.bld.yaml` file.

```yaml
name: "bld"             # Name of the target.

sources:
- name: <name>          # Name of the source (will be used in import blocks).
  target: <directory>   # Directory to import.
  files:
  - <filename>          # List of files, if set only these files will be used.

volumes:
- name: <name>          # Name of the volume.
  target: <directory>   # Path on the host filesystem.

steps:
- name: <name>          # Step name must be unique within the project.
  image: <image>        # Docker image.
  commands:
  - <cmd>               # Commands to run as part of the container entrypoint.
  workdir: <dir>        # Working directory on the container filesystem.
  user: <user>          # Docker user.
  env:
  - <KEY>=<VAL>         # Environment variables.
  volumes:
  - source: <name>      # Volume name, will not trigger a rebuild if changed.
    mount: <directory>  # Container filesystem mount point.
  imports:
  - source: <name>      # Source name, will trigger a rebuild if changed.
    mount: <directory>  # Container filesystem mount point.
  exports:
  - source: <name>      # New source name, can be imported by other steps.
    mount: <directory>  # Container filesystem mount point.
  # Build will commit the provided image and save the state of this current
  # image to the registry.
  build:
    tag: bld/example    # Local image tag.
    entrypoint: []      # Committed image entrypoint.
    command: []         # Committed image command.
    env: []             # Committed image environment.
    workdir:            # Committed image working directory.
    user:               # User for the docker image.
```
