package cmd

import (
	"github.com/docker/docker/client"
	"github.com/jessevdk/go-flags"
	"context"
	"log"
	"github.com/docker/docker/api/types"
)

type StopCommand struct {
	ctx context.Context
	cli *client.Client
}

func RegisterStopCommand(parser *flags.Parser) *StopCommand {
	cmd := &StopCommand{}
	parser.AddCommand("stop", "stops example galera cluster", "", cmd)
	return cmd
}

func (cmd *StopCommand) stopContainers() {
	containers, err := cmd.cli.ContainerList(cmd.ctx, types.ContainerListOptions{})
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

	nodes = append(nodes, firstNode)

	for _, container := range nodes {
		log.Print("Stopping container ", container.ID[:10], "... ")
		if err := cmd.cli.ContainerStop(cmd.ctx, container.ID, nil); err != nil {
			panic(err)
		}
		log.Println("Success")
	}
}

func (cmd *StopCommand) Execute(args []string) error {
	var err error
	cmd.ctx = context.Background()
	cmd.cli, err = client.NewEnvClient()
	if err != nil {
		log.Fatal(err, " :unable to init client")
	}

	cmd.stopContainers()

	return nil
}

