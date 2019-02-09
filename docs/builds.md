# Builds

This describes the behaviour of the build tool step by step.

Initializaton:

1. Submit the build to the graph module.
  - Graph module handles dependency resolution between the items. using a
    breadth first traversal to select which steps to run.
2. Start workers to handle steps to run.

Worker loop:

- Receive a unit of work from the graph:
- If it's a source, copy the source to a workspace directory that is inside
  the build directory. Also digest the given source and save this digest.
- If it's a step: run the step execution.

Step execution:

- Digest all imports + the step configuration.
- Check if a digest exists from the current cache.
  - If the digest exists:
    - Restore all exports from the found digest.
    - Restore a container if it exists.
  - If the digest does not exist:
    - Prepare exports for mounting.
    - Run the step and save all exports as new sources.
    - Push a container if `build` is present.

## Graph

The graph module tracks a directed acyclic graph. It maps steps and sources,
with a source being a dependency of a step. A step is only executed when all of
it's imports are marked as completed within the graph module.

Workers are popping work off of the graph and blocking until work is available.
