package main

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/container"
	"fmt"
	"github.com/jessevdk/go-flags"
	"galera-manager/internal/cmd"
)

// docker run --detach=true --name node1 -h node1 galera-test --wsrep-cluster-name=local-test --wsrep-cluster-address=gcomm://
// docker run --detach=true --name node2 -h node2 --link node1:node1 galera-test --wsrep-cluster-name=local-test --wsrep-cluster-address=gcomm://node1
// docker run --detach=true --name node3 -h node3 --link node1:node1 galera-test --wsrep-cluster-name=local-test --wsrep-cluster-address=gcomm://node1
// docker exec -ti node1 mysql -e 'show status like "wsrep_cluster_size"'
func main() {
	parser := flags.NewParser(nil, flags.Default)
	cmd.RegisterInstallCommand(parser)
	cmd.RegisterStartCommand(parser)
	cmd.RegisterStopCommand(parser)
	cmd.RegisterStatusCommand(parser)
	cmd.RegisterUninstallCommand(parser)
	cmd.RegisterHttpServerCommand(parser)

	parser.CommandHandler = func(command flags.Commander, args []string) error {
		//v := reflect.ValueOf(command)
		//logLevelValue := v.Elem().FieldByName("LogLevel")
		//if !logLevelValue.IsValid() {
		//	panic("reflection for LogLevel failed")
		//}
		//logFormatValue := v.Elem().FieldByName("LogFormat")
		//if !logFormatValue.IsValid() {
		//	panic("reflection for LogFormat failed")
		//}
		//
		//logLevel, _ := logrus.ParseLevel(logLevelValue.String())
		//logging.SetLogLevel(logLevel)
		//logging.SetFormat(logFormatValue.String())
		//
		//start := time.Now()

		err := command.Execute(args)

		//logging.GetLogger().Infof("Took %fs to run the command", time.Since(start).Seconds())

		return err
	}

	parser.Parse()
	return

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err, " :unable to init client")
	}

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
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		log.Fatal(err, " :unable to write tar header")
	}
	_, err = tw.Write(doclerFileContent)
	if err != nil {
		log.Fatal(err, " :unable to write tar body")
	}
	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	imageBuildResponse, err := cli.ImageBuild(
		ctx,
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

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "galera-test",
		Cmd:   []string{"--wsrep-cluster-name=local-test", "--wsrep-cluster-address=gcomm://"},
		Tty:   true,
	}, nil, nil, "node1")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(resp.ID)

	resp, err = cli.ContainerCreate(ctx, &container.Config{
		Image: "galera-test",
		Cmd:   []string{"--wsrep-cluster-name=local-test", "--wsrep-cluster-address=gcomm://"},
		Tty:   true,
	}, nil, nil, "node2")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(resp.ID)

	resp, err = cli.ContainerCreate(ctx, &container.Config{
		Image: "galera-test",
		Cmd:   []string{"--wsrep-cluster-name=local-test", "--wsrep-cluster-address=gcomm://"},
		Tty:   true,
	}, nil, nil, "node3")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(resp.ID)

	//_, err = cli.ContainerWait(ctx, resp.ID)
	//if err != nil {
	//	panic(err)
	//}
	//
	//out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	//if err != nil {
	//	panic(err)
	//}
	//
	//io.Copy(os.Stdout, out)
}