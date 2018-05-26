package cmd

import (
	"github.com/docker/docker/client"
	"github.com/jessevdk/go-flags"
	"log"
	"github.com/docker/docker/api/types"
	"context"
)

type UninstallCommand struct {
	ctx context.Context
	cli *client.Client
}

func RegisterUninstallCommand(parser *flags.Parser) *UninstallCommand {
	cmd := &UninstallCommand{}
	parser.AddCommand("uninstall", "uninstalls example galera cluster", "", cmd)
	return cmd
}

func (cmd *UninstallCommand) stopAndRemoveContainers() {
	containers, err := cmd.cli.ContainerList(cmd.ctx, types.ContainerListOptions{All:true})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		if _, ok := container.Labels["galera-test"]; !ok {
			continue
		}
		//
		//log.Print("Stopping container ", container.ID[:10], "... ")
		//if err := cmd.cli.ContainerStop(cmd.ctx, container.ID, nil); err != nil {
		//	panic(err)
		//}
		//log.Println("Success")

		log.Print("Removing container ", container.ID[:10], "... ")
		if err := cmd.cli.ContainerRemove(cmd.ctx, container.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
			panic(err)
		}
		log.Println("Success")
	}
}

func (cmd *UninstallCommand) removeNetwork(networkName string) {
	networks, err := cmd.cli.NetworkList(cmd.ctx, types.NetworkListOptions{})
	if err != nil {
		panic(err)
	}

	for _, network := range networks {
		if network.Name != networkName {
			continue
		}

		log.Print("Removing network ", network.ID[:10], "... ")
		if err := cmd.cli.NetworkRemove(cmd.ctx, network.ID); err != nil {
			log.Fatalf("%v :unable to remove network %s", err, network.ID)
		}
		log.Println("Success")
	}
}

func (cmd *UninstallCommand) removeImage(image string) {
	images, err := cmd.cli.ImageList(cmd.ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, v := range images {
		if !containsString(v.RepoTags, image) {
			continue
		}

		log.Print("Removing image ", v.ID[:10], "... ")
		if _, err = cmd.cli.ImageRemove(cmd.ctx, v.ID, types.ImageRemoveOptions{Force: true, PruneChildren: true}); err != nil {
			log.Fatalf("%v :unable to remove image %s", err, v.ID)
		}
		log.Println("Success")
	}
}

func (cmd *UninstallCommand) Execute(args []string) error {
	var err error
	cmd.ctx = context.Background()
	cmd.cli, err = client.NewEnvClient()
	if err != nil {
		log.Fatal(err, " :unable to init client")
	}

	cmd.stopAndRemoveContainers()
	cmd.removeNetwork("galera-test-network")
	cmd.removeImage("galera-test:latest")

	return nil
}
