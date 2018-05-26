package cmd

import (
	"github.com/docker/docker/client"
	"context"
	"log"
	"github.com/docker/docker/api/types"
	"github.com/jessevdk/go-flags"
)

type StartCommand struct {
	ctx context.Context
	cli *client.Client
}

func RegisterStartCommand(parser *flags.Parser) *StartCommand {
	cmd := &StartCommand{}
	parser.AddCommand("start", "starts example galera cluster", "", cmd)
	return cmd
}


func (cmd *StartCommand) startContainer(containerId string) {
	if err := cmd.cli.ContainerStart(cmd.ctx, containerId, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

	func (cmd *StartCommand) startContainers() {
		containers, err := cmd.cli.ContainerList(cmd.ctx, types.ContainerListOptions{All:true})
		if err != nil {
			panic(err)
		}

		var firstNode types.Container
		nodes := make([]types.Container, 0)

		for _, container := range containers {
			if _, ok := container.Labels["galera-test"]; !ok {
				continue
			}

			if firstNodeValue, ok := container.Labels["galera-test-first-node"]; ok && firstNodeValue == "1" {
				firstNode = container
				continue
			}

			nodes = append(nodes, container)
		}

		if firstNode.ID == "" {
			panic("First node not found")
		}

		nodes = append([]types.Container{firstNode}, nodes...)

		for _, container := range nodes {
			if container.State != "created" && container.State != "exited" {
				log.Printf("Not starting container %v as it's now in state %s\n", container.Names, container.State)
				continue
			}

			cmd.startContainer(container.ID)
			log.Printf("Container %v started\n", container.Names)
		}
	}

func (cmd *StartCommand) Execute(args []string) error {
	var err error
	cmd.ctx = context.Background()
	cmd.cli, err = client.NewEnvClient()
	if err != nil {
		log.Fatal(err, " :unable to init client")
	}

	cmd.startContainers()

	return nil
}
