package cmd

import (
	"github.com/docker/docker/client"
	"github.com/jessevdk/go-flags"
	"context"
	"log"
	"net/http"
	"galera-cluster/internal/cluster"
	"fmt"
	"github.com/docker/docker/api/types"
)

type HttpServerCommand struct {
	ctx context.Context
	cli *client.Client
}

func RegisterHttpServerCommand(parser *flags.Parser) *HttpServerCommand {
	cmd := &HttpServerCommand{}
	parser.AddCommand("http-server", "starts http server to monitor the cluster state", "", cmd)
	return cmd
}

func (cmd *HttpServerCommand) getMetric(activeNode types.Container, metric string) string {
	value, err := cluster.GetNodeMetric(cmd.ctx, cmd.cli, activeNode, metric)
	if err != nil {
		return fmt.Sprintf("Unable to obtain metric %s\n", metric)
	}

	return fmt.Sprintf("%s: %s\n", metric, value)
}

func (cmd *HttpServerCommand) getHandler() (func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		cmd.ctx = context.Background()
		cmd.cli, err = client.NewEnvClient()
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprint(err, " :unable to init client")))
			return
		}

		activeNode, nodes, err := cluster.GetContainers(cmd.ctx, cmd.cli)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprint(err, " :unable to get containers")))
			return
		}

		var output string

		for _, v := range nodes {
			output += fmt.Sprintf("Node %v has status %s\n", v.Names, v.Status)
		}

		if activeNode.ID == "" {
			w.Write([]byte(output + "No node is running"))
			return
		}

		output += cmd.getMetric(activeNode, "wsrep_provider_name")
		output += cmd.getMetric(activeNode, "wsrep_ready")
		output += cmd.getMetric(activeNode, "wsrep_cluster_size")

		w.Write([]byte(output))
	}
}

func (cmd *HttpServerCommand) runServer() {
	http.HandleFunc("/", cmd.getHandler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (cmd *HttpServerCommand) Execute(args []string) error {
	var err error
	cmd.ctx = context.Background()
	cmd.cli, err = client.NewEnvClient()
	if err != nil {
		log.Fatal(err, " :unable to init client")
	}

	cmd.runServer()

	return nil
}

