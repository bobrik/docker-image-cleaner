package main

import (
	"flag"
	"github.com/fsouza/go-dockerclient"
	"log"
	"strings"
)

func main() {
	endpoint := flag.String("endpoint", "unix:///var/run/docker.sock", "docker api endpoint")
	dryRun := flag.Bool("dry-run", false, "just list containers to remove")
	flag.Parse()

	client, err := docker.NewClient(*endpoint)
	if err != nil {
		log.Fatal(err)
	}

	images, err := client.ListImages(false)
	if err != nil {
		log.Fatal(err)
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.Fatal(err)
	}

	used := map[string]string{}

	for _, container := range containers {
		inspected, err := client.InspectContainer(container.ID)
		if err != nil {
			log.Println("error getting container info for "+container.ID, err)
			continue
		}

		used[inspected.Image] = container.ID
	}

	for _, image := range images {
		if _, ok := used[image.ID]; !ok {
			log.Printf("Going to remove %s: %s", image.ID, strings.Join(image.RepoTags, ","))

			repos := map[string]bool{}
			for _, tag := range image.RepoTags {
				d := strings.Index(tag, "/")
				if d != -1 {
					repos[tag[0:d]] = true
				}
			}

			remove := []string{}
			if len(repos) > 1 {
				remove = image.RepoTags
			} else {
				remove = append(remove, image.ID)
			}

			if !*dryRun {
				for _, r := range remove {
					err := client.RemoveImage(r)
					if err != nil {
						log.Printf("error while removing %s (%s): %s", r, strings.Join(image.RepoTags, ","), err)
						continue
					}
				}
			}
		}
	}
}
