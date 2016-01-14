package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

func main() {
	exclude := flag.String("exclude", "", "images to exclude, image:tag[,image:tag]")
	dryRun := flag.Bool("dry-run", false, "just list containers to remove")
	flag.Parse()

	if os.Getenv("DOCKER_HOST") == "" {
		err := os.Setenv("DOCKER_HOST", "unix:///var/run/docker.sock")
		if err != nil {
			log.Fatalf("error setting default DOCKER_HOST: %s", err)
		}
	}

	excluded := map[string]struct{}{}

	if len(*exclude) > 0 {
		for _, i := range strings.Split(*exclude, ",") {
			excluded[i] = struct{}{}
		}
	}

	docker, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("error creating docker client: %s", err)
	}

	topImages, err := docker.ImageList(types.ImageListOptions{})
	if err != nil {
		log.Fatalf("error getting docker images: %s", err)
	}

	allImages, err := docker.ImageList(types.ImageListOptions{All: true})
	if err != nil {
		log.Fatalf("error getting all docker images: %s", err)
	}

	imageTree := make(map[string]types.Image, len(allImages))
	for _, image := range allImages {
		imageTree[image.ID] = image
	}

	containers, err := docker.ContainerList(types.ContainerListOptions{All: true})
	if err != nil {
		log.Fatalf("error getting docker containers: %s", err)
	}

	used := map[string]string{}

	for _, container := range containers {
		inspected, err := docker.ContainerInspect(container.ID)
		if err != nil {
			log.Printf("error getting container info for %s: %s", container.ID, err)
			continue
		}

		used[inspected.Image] = container.ID

		parent := imageTree[inspected.Image].ParentID
		for {
			if parent == "" {
				break
			}

			used[parent] = container.ID
			parent = imageTree[parent].ParentID
		}
	}

	for _, image := range topImages {
		if _, ok := used[image.ID]; !ok {
			skip := false
			for _, tag := range image.RepoTags {
				if _, ok := excluded[tag]; ok {
					skip = true
				}

				if skip {
					break
				}
			}

			if skip {
				log.Printf("Skipping %s: %s", image.ID, strings.Join(image.RepoTags, ","))
				continue
			}

			log.Printf("Going to remove %s: %s", image.ID, strings.Join(image.RepoTags, ","))

			if !*dryRun {
				if len(image.RepoTags) < 2 {
					// <none>:<none> case, just remove by id
					_, err := docker.ImageRemove(types.ImageRemoveOptions{ImageID: image.ID, PruneChildren: true})
					if err != nil {
						log.Printf("error while removing %s (%s): %s", image.ID, strings.Join(image.RepoTags, ","), err)
					}
				} else {
					// several tags case, remove each by name
					for _, r := range image.RepoTags {
						_, err := docker.ImageRemove(types.ImageRemoveOptions{ImageID: r, PruneChildren: true})
						if err != nil {
							log.Printf("error while removing %s (%s): %s", r, strings.Join(image.RepoTags, ","), err)
							continue
						}
					}
				}
			}
		}
	}
}
