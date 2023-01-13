# zru - Zenith Release Utility

`zru` is the toolbox used to deploy, maintain, and sync Zenith's projects across repositories.

## TLDR (too long, didn't read)

```bash
# build the app
make

# display the help prompt
./zru -h

# sync across repositories
./zru release --path ./repository --remote backup --branch backup-$(date +%s) --src-branch develop

# deploy
./zru deploy --url https://some-url.com --path repository

# request SSL certs on first deployment
./zru certs --email myemail@email.com --url example.com
```

> You can also run `make help` to check makefile rules

## Installation

Compile like so:

```bash
make
```

Install like so:
```bash
# assuming you GOPATH is set
make install
```

### Build a Docker image

```bash
make docker

# or:
docker build . -t zru
```

## Usage

### clone

Clones a repository. Mainly used for CI, but could be useful for you too.

```bash
clones a repository recursively and print latest commit hash

Usage:
  clone [flags]
  clone [command]

Available Commands:
  certs       fetches new SSL certificates
  completion  Generate the autocompletion script for the specified shell
  deploy      deploys the Zenith timesheet tool
  help        Help about any command
  nuke        nukes docker volumes, images, & everything
  release     delivers code release across repositories

Flags:
  -h, --help          help for clone
  -p, --path string   repository path (default ".")
  -u, --url string    git clone url (default "https://gitlab.com/zenith-hr/TIMESHEET.git")
  -v, --version       version for clone

Use "clone [command] --help" for more information about a command.
```

Example:
```bash
zru clone --url git@gitlab.com:bogdzn/zru.git --path /tmp/zru
```

### nuke

Deletes docker-compose setup from existence. **Use only on development servers**

```bash
destroys all deployment config for dev purposes

Usage:
  clone nuke [flags]

Flags:
  -h, --help      help for nuke
  -v, --version   version for nuke
```

Example:
```bash
zru nuke
```

### certs

Generates SSL certificates for a new environment. Use this before deployment to avoid an `nginx` crash.

```bash
To be used on initial deployments -- generates new SSL certificates for your new env

Usage:
  clone certs [flags]

Flags:
  -e, --email string   renewal email (default "backoffice@zenith-rh.com")
  -h, --help           help for certs
  -u, --url string     domain of your new environment (default "timesheet.zenith-rh.com")
  -v, --version        version for certs
```

Example:
```bash
zru certs --email myemail@email.com --url example.com
```

### deploy

Renames all occurences of `oldUrl` to the correct url for the environment, then pulls up the infrastructure through docker-compose.

```bash
wraps around docker-compose to deploy the various services

Usage:
  clone deploy [flags]

Flags:
  -h, --help                  help for deploy
  -o, --original-url string   old deploy URL (default "timesheet.zenith-rh.com")
  -p, --path string           repository path (default ".")
  -u, --url string            new deploy URL (default "qa-timesheet.zenith-rh.com")
  -v, --version               version for deploy
```

Example:
```bash
zru deploy --url some-url.com --path repository
```

### release

Prepares a release from repository to repository.

```bash
creates target branch from source branch & pushes it to new remote/branch for software releases

Usage:
  release [flags]
  release [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command

Flags:
  -h, --help                   help for release
  -p, --path string            repository path (default ".")
  -e, --remote string          git remote target (default "customer")
  -b, --src-branch string      branch from which we deliver release (default "master")
  -r, --src-remote string      git remote source (default "origin")
  -t, --target-branch string   branch to deliver release to (default "release")
  -v, --version                version for release

Use "release [command] --help" for more information about a command.
```

Example:
```bash
zru release --path ./repository --remote backup --branch backup-$(date +%s) --src-branch develop
```
