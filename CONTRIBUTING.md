# Contributing to Geek Swimmers

## Configuration

### User Session

## IDEs

### VSCode

#### Debugging

For those who use VSCode to contribute to Geek Swimmers, be aware that the [`.vscode`](https://github.com/htmfilho/geekswimmers/tree/main/.vscode) folder, pushed to the repository, contains the file [`launch.json`](https://github.com/htmfilho/geekswimmers/blob/main/.vscode/launch.json) that configures the debugger to work properly on this code editor. Put some breakpoints around and check it out!

#### Development Containers

Development Containers are used to provide a development environment without having to setup the local machine. You install all your development tools into the container, and VSCode will run the container and connect to it. This allows for a consistent environment among developers.

Ref: https://containers.dev/
VSCode plugin: https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers

### Howto

Once you install the extension, VSCode will notice you have the `.devcontainer` folder in the project, and ask to reopen in the `devcontainer`. If you don't have the docker environment built, it will build the container, then run it and connect you.

## Continuous Integration

We use Github Actions as continuous integration service.

### Go Linter

The linter is configured to run on every push and pull request, verifying the quality of the code. More info: https://github.com/golangci/golangci-lint-action

## Hosting

### Heroku

    $ heroku ps:scale web=1