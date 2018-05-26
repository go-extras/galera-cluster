package cmd

import (
	"github.com/jessevdk/go-flags"
	"archive/tar"
	"log"
	"bytes"
	"github.com/docker/docker/api/types"
	"io"
	"os"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/network"
)

type InstallCommand struct {
	ctx context.Context
	cli *client.Client
}

func RegisterInstallCommand(parser *flags.Parser) *InstallCommand {
	cmd := &InstallCommand{}
	parser.AddCommand("install", "installs example galera cluster", "", cmd)
	return cmd
}

func (cmd *InstallCommand) buildImage() {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	dockerFile := "Dockerfile"
	doclerFileContent := []byte(`
FROM ubuntu:16.04
ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update
RUN apt-get install -y software-properties-common
RUN apt-key adv --recv-keys --keyserver hkp://keyserver.ubuntu.com:80 BC19DDBA
RUN add-apt-repository 'deb http://releases.galeracluster.com/mysql-wsrep-5.7/ubuntu xenial main'
RUN add-apt-repository 'deb http://releases.galeracluster.com/galera-3/ubuntu xenial main'
RUN apt-get update
RUN apt-get install -y galera-3 galera-arbitrator-3 mysql-wsrep-5.7 rsync lsof
RUN echo -n "[mysqld]\n\
user = mysql\n\
bind-address = 0.0.0.0\n\
wsrep_provider = /usr/lib/galera/libgalera_smm.so\n\
wsrep_sst_method = rsync\n\
default_storage_engine = innodb\n\
binlog_format = row\n\
innodb_autoinc_lock_mode = 2\n\
innodb_flush_log_at_trx_commit = 0\n\
query_cache_size = 0\n\
query_cache_type = 0\n" > /etc/mysql/my.cnf
RUN mkdir /var/run/mysqld
RUN chown mysql:mysql /var/run/mysqld
ENTRYPOINT ["mysqld"]
`)

	tarHeader := &tar.Header{
		Name: dockerFile,
		Size: int64(len(doclerFileContent)),
	}
	err := tw.WriteHeader(tarHeader)
	if err != nil {
		log.Fatal(err, " :unable to write tar header")
	}
	_, err = tw.Write(doclerFileContent)
	if err != nil {
		log.Fatal(err, " :unable to write tar body")
	}
	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	imageBuildResponse, err := cmd.cli.ImageBuild(
		cmd.ctx,
		dockerFileTarReader,
		types.ImageBuildOptions{
			Tags: []string{"galera-test"},
			Context:    dockerFileTarReader,
			Dockerfile: dockerFile,
			Remove:     true,
		})
	if err != nil {
		log.Fatal(err, " :unable to build docker image")
	}
	defer imageBuildResponse.Body.Close()
	_, err = io.Copy(os.Stdout, imageBuildResponse.Body)
	if err != nil {
		log.Fatal(err, " :unable to read image build response")
	}
}

func (cmd *InstallCommand) createNodeContainer(containerName, linkTo string) {
	var firstNode string

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"galera-test-network": {
				Aliases: []string{"galera-test-network"},
			},
		},
	}

	if linkTo == "" {
		firstNode = "1"
	}

	resp, err := cmd.cli.ContainerCreate(cmd.ctx, &container.Config{
		Image: "galera-test",
		Cmd:   []string{"--wsrep-cluster-name=local-test", "--wsrep-cluster-address=gcomm://"+linkTo},
		Tty:   true,
		Labels: map[string]string{
			"galera-test-first-node": firstNode,
			"galera-test": containerName,
		},
	}, nil, networkingConfig, containerName)
	if err != nil {
		log.Fatal(err, " :unable to create a container")
	}
	fmt.Println(resp.ID)
}

func (cmd *InstallCommand) createNetwork(networkName string) {
	if _, err := cmd.cli.NetworkCreate(cmd.ctx, networkName, types.NetworkCreate{}); err != nil {
		log.Fatal(err, " :unable to create a network")
	}
}

func (cmd *InstallCommand) Execute(args []string) error {
	var err error
	cmd.ctx = context.Background()
	cmd.cli, err = client.NewEnvClient()
	if err != nil {
		log.Fatal(err, " :unable to init client")
	}

	cmd.buildImage()

	cmd.createNetwork("galera-test-network")

	cmd.createNodeContainer("node1", "")
	cmd.createNodeContainer("node2", "node1")
	cmd.createNodeContainer("node3", "node1")

	return nil
}
