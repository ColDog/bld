# bld

Monorepo build tool. With change detection.

## Structure

- `Source`: A folder.
- `Step`: A set of commands to execute inside a container.
- `Service`: A container to be run alongside the steps.


1. For a local source, we just checksum insert into the source map.
2. For a remote source created during a previous build we want to load the
   source for a given build digest. Restore the local state to the previous state.
