# Contributing to Geek Swimmers

## Getting It Up and Running

    $ go run main.go

## Tests

To run all tests:

    $ go test ./...

To run the tests of an specific package:

    $ go test geekswimmers/utils

## Upgrading Go

To upgrade the version of Go, we have to increment it in the following files:

* `go.mod`
* `Dockerfile`
* `.github/workflows/go.yml`
* `.github/workflows/golangcl-int.yml`

We may need to increment the version of the lint too at `.github/workflows/golangcl-int.yml`, keeping up with the evolution of the language.

## IDEs

### VSCode

#### Debugging

For those who use VSCode to contribute to Geek Swimmers, be aware that the [`.vscode`](https://github.com/htmfilho/geekswimmers/tree/main/.vscode) folder, pushed to the repository, contains the file [`launch.json`](https://github.com/htmfilho/geekswimmers/blob/main/.vscode/launch.json) that configures the debugger to work properly on this code editor. Put some breakpoints around and check it out!

#### Development Containers

Development Containers are used to provide a development environment without having to setup the local machine. You install all your development tools into the container, and VSCode will run the container and connect to it. This allows for a consistent environment among developers.

Ref: https://containers.dev/
VSCode plugin: https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers

Once you install the extension, VSCode will notice you have the `.devcontainer` folder in the project, and ask to reopen in the `devcontainer`. If you don't have the docker environment built, it will build the container, then run it and connect you.

## Continuous Integration

We use Github Actions as continuous integration service.

### Go Linter

The linter is configured to run on every push and pull request, verifying the quality of the code. More info: https://github.com/golangci/golangci-lint-action

It is recommended to run a linter locally before pushing the branch to the server. To do it, run:

    $ golangci-lint run

## Hosting

### Heroku

    $ heroku ps:scale web=1
