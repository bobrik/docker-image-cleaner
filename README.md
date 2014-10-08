# Docker image cleaner

The long missing command to clean up all docker images that are not used by any container.

## Usage

This command is available as docker image `bobrik/image-cleaner`:

```
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock bobrik/image-cleaner
```

Add `-dry-run` to the end if you want to see what is going to be deleted.
