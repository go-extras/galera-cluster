package cmd

import (
	"github.com/docker/docker/client"
	"github.com/jessevdk/go-flags"
	"context"
	"log"
	"galera-manager/internal/cluster"
	"fmt"
	"github.com/docker/docker/api/types"
)

type StatusCommand struct {
	ctx context.Context
	cli *client.Client
}

func RegisterStatusCommand(parser *flags.Parser) *StatusCommand {
	cmd := &StatusCommand{}
	parser.AddCommand("status", "shows example galera cluster status", "", cmd)
	return cmd
}

func (cmd *StatusCommand) printMetric(activeNode types.Container, metric string) {
	value, err := cluster.GetNodeMetric(cmd.ctx, cmd.cli, activeNode, metric)
	if err != nil {
		log.Fatal(err, " :unable to obtain metric", metric)
	}
	fmt.Println(metric, value)
}

func (cmd *StatusCommand) Execute(args []string) error {
	var err error
	cmd.ctx = context.Background()
	cmd.cli, err = client.NewEnvClient()
	if err != nil {
		log.Fatal(err, " :unable to init client")
	}

	activeNode, nodes, err := cluster.GetContainers(cmd.ctx, cmd.cli)
	if err != nil {
		log.Fatal(err, " :unable to get containers")
	}

	for _, v := range nodes {
		fmt.Printf("Node %v has status %s\n", v.Names, v.Status)
	}

	if activeNode.ID == "" {
		fmt.Println("No node is running")
		return nil
	}

	cmd.printMetric(activeNode, "wsrep_provider_name")
	cmd.printMetric(activeNode, "wsrep_ready")
	cmd.printMetric(activeNode, "wsrep_cluster_size")

	return nil
}

