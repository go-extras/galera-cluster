package cluster

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"io/ioutil"
	"strings"
	"context"
	"github.com/docker/docker/client"
	"log"
	"errors"
)

//printStatus(activeNode, "wsrep_provider_name")
//printStatus(activeNode, "wsrep_ready")
//printStatus(activeNode, "wsrep_cluster_size")
func GetNodeMetric(ctx context.Context, cli *client.Client, activeNode types.Container, metric string) (string, error) {
	execOpts := types.ExecConfig{
		Cmd: []string{"mysql", "-e", fmt.Sprintf("show status like \"%s\"", metric)},
		AttachStdout: true,
		//Tty: true,
	}

	execInstance, err := cli.ContainerExecCreate(ctx, activeNode.ID, execOpts)
	if err != nil {
		return "", err
	}
	att, err := cli.ContainerExecAttach(ctx, execInstance.ID, execOpts)
	if err != nil {
		return "", err
	}
	execStartOpts := types.ExecStartCheck{
		Detach: false,
		Tty: false,
	}

	if err = cli.ContainerExecStart(ctx, execInstance.ID, execStartOpts); err != nil {
		return "", err
	}

	if _, err = cli.ContainerExecInspect(ctx, execInstance.ID); err != nil {
		return "", err
	}

	result, err := ioutil.ReadAll(att.Reader)
	if err != nil {
		return "", err
	}

	response := strings.Split(string(result), "\n")
	if len(response) < 2 {
		return "", errors.New(fmt.Sprintf("cannot obtain %s", metric))
	}

	response = strings.Split(response[1], "\t")
	if len(response) < 2 {
		return "", errors.New(fmt.Sprintf("cannot obtain %s", metric))
	}

	return response[1], nil
}

func GetContainers(ctx context.Context, cli *client.Client) (firstActiveNode types.Container, nodes []types.Container, err error) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}

	firstNode := types.Container{}
	nodes = make([]types.Container, 0)

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

	if firstNode.ID != "" {
		log.Printf("First node (%v) container ID: %s, status: %s", firstNode.Names, firstNode.ID[:10], firstNode.Status)
	} else {
		log.Println("First node not found")
	}

	if len(nodes) > 0 {
		for _, container := range nodes {
			log.Printf("Node %v container ID: %s, status: %s", container.Names, container.ID[:10], container.Status)
		}
	} else {
		log.Println("No other nodes found")
	}

	if firstNode.ID == "" && len(nodes) == 0 {
		return firstActiveNode, nil, errors.New("no nodes found")
	}

	allNodes := make([]types.Container, 0)
	if firstNode.ID != "" {
		allNodes = append([]types.Container{firstNode}, nodes...)
	} else {
		allNodes = nodes[:]
	}

	for _, v := range allNodes {
		if v.State == "running" {
			firstActiveNode = v
			break
		}
	}

	return firstActiveNode, allNodes, nil
}

