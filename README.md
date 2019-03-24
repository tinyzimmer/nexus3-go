# nexus3-go

[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=round-square)](http://godoc.org/github.com/tinyzimmer/nexus3-go)

`nexus3-go` provides a golang client and methods for the Sonatype Nexus3 API. I aim to have all the REST
methods implemented, but the main feature is the script interface that provides for functionality
that does not exist via the external API.

I'll be sure to spruce up this README if I continue this project, but for documentation for now please consult
[GoDoc](http://godoc.org/github.com/tinyzimmer/nexus3-go).

## nexus3-go Command

The command is more for showcasing and testing the underlying API methods, however it does provide some convenient functions.
It's `kingpin` also which allows for automatic shell completion among other benefits. To build the command, run:

```bash
$> make build
```

```bash
$> bin/nexus-cmd
usage: bin/nexus-cmd [<flags>] <command> [<args> ...]

A command-line interface for Sonatype Nexus 3.

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
  -h, --host="http://localhost:8081"
                             URL of the Nexus host
  -u, --username="admin"     Username to authenticate to Nexus
  -p, --password="admin123"  Password to authenticate to Nexus

Commands:
  help [<command>...]
    Show help.

  groovy-exec [<flags>] [<commands>...]
    Execute a groovy script on the Nexus host

  list-repositories
    List the repositories in Nexus

  list-blob-stores
    List the blob stores in Nexus

  list-formats
    List the available component formats

  list-assets <repository>
    List the assets for a given repository

  list-components <repository>
    List the components for a given repository

  upload-component --repository=REPOSITORY --type=TYPE --file=FILE
    Upload a component to a given repository (does not support maven uploads at the moment)

  create-blobstore --name=NAME [<flags>]
    Create a new blob store

  delete-blobstore [<flags>] [<blobstore>]
    Delete a blobstore by the given name

```
