# Docker image cleaner

The long missing command to clean up all docker images that are not used by any container.

## Using

This command is available as docker image `bobrik/image-cleaner`:

```
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock bobrik/image-cleaner
```
