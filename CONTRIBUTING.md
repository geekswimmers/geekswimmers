# Contributing to Geek Swimmers

## IDEs

### VSCode

For those who use VSCode to contribute to Geek Swimmers, be aware that the [`.vscode`](https://github.com/htmfilho/geekswimmers/tree/main/.vscode) folder, pushed to the repository, contains the file [`launch.json`](https://github.com/htmfilho/geekswimmers/blob/main/.vscode/launch.json) that configures the debugger to work properly on this code editor. Put some breakpoints around and check it out!

## Development Containers

Development Containers are used to provide a development envrionment without having to setup the local machine - you install all your development tools into the container, and vscode will run the container and 
connect to it.  This allows for a consistent environment between developers.

Ref: https://containers.dev/
vscode plugin: https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers

### Howto

Once you install the extension, when you open the project vscode will notice you have the .devcontainer folder, and ask to reopen in the devcontainer.
If you don't have the docker envt built, it will build the container, then run it and connect you.
