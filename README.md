# Tacher
## _Start your new Spring Boot project <sup>with Go</sup>_

Tacher lets you configure a Spring Boot project without the need to deal with zip files.  
You can set up your project with an interface similar to [Spring Initializr](https://start.spring.io/) and you'll find the new project in the folder you want, without the need to download and extract any zip archive.

## Prerequisites
The only prerequisite is Go `>=1.19`

## Build
You can build Tacher by running  

```bash
go build ./src
```

Then you will find the `tacher` executable in the directory you're in.

## Execute
Running `tacher` without any argument will print an overview of the available commands.  

To start to generate your new project you have to run `./tacher init`, then follow the wizard.
