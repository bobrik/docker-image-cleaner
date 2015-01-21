package main

import (
	"flag"
	"github.com/fsouza/go-dockerclient"
	"log"
	"strings"
)

func main() {
	endpoint := flag.String("endpoint", "unix:///var/run/docker.sock", "docker api endpoint")
	exclude := flag.String("exclude", "", "images to exclude, image:tag[,image:tag]")
	dryRun := flag.Bool("dry-run", false, "just list containers to remove")
	flag.Parse()

	excluded := map[string]struct{}{}

	if len(*exclude) > 0 {
		for _, i := range strings.Split(*exclude, ",") {
			excluded[i] = struct{}{}
		}
	}

	client, err := docker.NewClient(*endpoint)
	if err != nil {
		log.Fatal(err)
	}

	images, err := client.ListImages(docker.ListImagesOptions{})
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

			repos := map[string]struct{}{}
			for _, tag := range image.RepoTags {
				d := strings.Index(tag, "/")
				if d != -1 {
					repos[tag[0:d]] = struct{}{}
				} else {
					repos["_"] = struct{}{}
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
