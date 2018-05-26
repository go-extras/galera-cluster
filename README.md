Galera Cluster Manager
======================

This utility let's you create and monitor MySQL Galera cluster.


## Compilation

### Prerequisites

This project requires Go 1.10 or newer. If you haven't installed Go yet, use the binary installers available from the [official Go website](https://golang.org/dl/).

### Compilation

- Install `dep`: https://github.com/golang/dep

```bash
go get -u github.com/golang/dep/cmd/dep
```

NB: if you cannot run `dep` after you installed it, try adding exporting the path:

```bash
export PATH=$PATH:$GOROOT/bin
```

- Set `GOPATH` to some directory:

```bash
export GOPATH=/home/joe/go-projects
```

- Clone the sources

```bash
mkdir -p /home/joe/go-projects/src
cd /home/joe/go-projects/src
git clone git@github.com:go-extras/galera-manager.git
cd  galera-manager
```

- Install dependencies:

```bash
dep ensure -vendor-only
```

- Run tests:

> NB: for now skip this step as no tests exist at the moment.

```bash
go test ./...
```

- Build the binary:

```bash
go build -o bin/galera-manager
```

After running the last command you will get a `./bin/galera-manager` executable.

## Configuration

This tool does not have any configuration file.

## How to use?

First of all, you must have Docker installed and running. For MacOS users you must use `"storage-driver" : "aufs"` in your config.

### Install the cluster

NB: on linux you might need to prefix the commands with `sudo`.

```
bash-3.2$ ./bin/galera-manager install
[... docker image build and container creation commands will be shown ...]
```

### Start the cluster

```
bash-3.2$ ./bin/galera-manager start
2018/05/26 12:11:44 Container [/node1] started
2018/05/26 12:11:45 Container [/node3] started
2018/05/26 12:11:45 Container [/node2] started
```

### Check cluster status

```
bash-3.2$ ./bin/galera-manager start
2018/05/26 12:11:44 Container [/node1] started
2018/05/26 12:11:45 Container [/node3] started
2018/05/26 12:11:45 Container [/node2] started
bash-3.2$ ./bin/galera-manager status
2018/05/26 12:12:20 First node ([/node1]) container ID: 0397a525f1, status: Up 36 seconds
2018/05/26 12:12:20 Node [/node3] container ID: 1a4260823f, status: Up 35 seconds
2018/05/26 12:12:20 Node [/node2] container ID: 30f90e228c, status: Up 34 seconds
Node [/node1] has status Up 36 seconds
Node [/node3] has status Up 35 seconds
Node [/node2] has status Up 34 seconds
wsrep_provider_name Galera
wsrep_ready ON
wsrep_cluster_size 3
```

### Start http server

```
./bin/galera-manager http-server
```

(no output will be shown)

The server will start bound to `:8080` and can be accessed at `http://localhost:8080`.

The output via http server will now be pretty much the same as via the console command:

```
bash-3.2$ curl http://localhost:8080
Node [/node1] has status Up 2 minutes
Node [/node3] has status Up 2 minutes
Node [/node2] has status Up 2 minutes
wsrep_provider_name: Galera
wsrep_ready: ON
wsrep_cluster_size: 3
```

### Stopping the nodes

```
bash-3.2$ ./galera-cluster stop
2018/05/26 12:16:31 Stopping container 1a4260823f...
2018/05/26 12:16:37 Success
2018/05/26 12:16:37 Stopping container 30f90e228c...
2018/05/26 12:16:47 Success
2018/05/26 12:16:47 Stopping container 0397a525f1...
2018/05/26 12:16:52 Success
```

Let's now check the status in our http endpoint:

```
bash-3.2$ curl http://localhost:8080
Node [/node1] has status Exited (0) 41 seconds ago
Node [/node3] has status Exited (0) 56 seconds ago
Node [/node2] has status Exited (137) 45 seconds ago
No node is running
```

### Uninstalling the cluster

```
bash-3.2$ ./galera-cluster uninstall
2018/05/26 12:18:16 Removing container 1a4260823f...
2018/05/26 12:18:16 Success
2018/05/26 12:18:16 Removing container 30f90e228c...
2018/05/26 12:18:16 Success
2018/05/26 12:18:16 Removing container 0397a525f1...
2018/05/26 12:18:16 Success
2018/05/26 12:18:16 Removing network af6160f3ad...
2018/05/26 12:18:16 Success
2018/05/26 12:18:16 Removing image sha256:eeb...
2018/05/26 12:18:17 Success
```

Now, let's check again the http endpoint:

```
bash-3.2$ curl http://localhost:8080
no nodes found :unable to get containers
```

## TODO

- Write tests
- Make http port configurable
- Make names (labels) configurable
- Better error handling
- Http output should be RESTful and there should be a JS UI for that
- Make it possible to expose MySQL ports
- Check cluster state via MySQL connection, not via command execution
- Check cluster state on all the nodes (optionally)
- Make purging child containers optional when uninstalling (for faster re-installs)
- Better logging facilities (use logrus library and error levels)