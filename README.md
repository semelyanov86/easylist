# EasyList
![Logo](https://sergeyem.ru/img/sprite.png)

Easy and best way to manage shopping and meal planning


## Tech Stack
**Client:** Vue.js 3, TailwindCSS

**Server:** Golang


## Installation

Clone project and generate binary, running make command

```bash
  make run
```

OR just download `api` executable from realeases folder

To fill database, first install migrate tool:
```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
mv migrate.linux-amd64 $GOPATH/bin/migrate
```

Before you continue, please check that it’s available and working on your machine by trying to execute the migrate binary with the -version flag. It should output the current version number similar to this:
```bash
$ migrate -version
4.14.1
```

To run migration, execute following command:
```bash
make migrate
```

## Environment Variables

To run this project, you will need to add the following environment variables to your .envrc file

`EASYLIST_DB_DSN=root:password@/easylist?parseTime=true`

To have a possibility to run integration tests, you need to set test DB DSN variable:
`EASYLIST_TEST_DB=root:pass@/easylist_test?parseTime=true&multiStatements=true`

## Command line arguments

You can run executable script with following arguments:

* `version` - Display script version and exit.

## Configuration file
There is 2 configuration example files.
`.envrc` - for storing environment variables
`config.yaml` - for storing app configuration

To create them, use .envrc.example and config.yaml.example files.
Put your config `easylist.yaml` file in `~/.config` directory

## Running Tests

To run tests, run the following command

```bash
  make audit
```


## Deployment

To deploy this project run

```bash
  production/deploy/api
```

## API Reference
You can find Insomnia yaml collection in folder `documentation/api`

## Features

- User registration and authorization via Bearer token
- List of folders and Lists
- Item storage with attachment. Each item links with 'List'
- Metrics and health check endpoints.



## Acknowledgements

- [Json:Api standard](https://jsonapi.org)

## Support

For support, email se@sergeyem.ru or via telegram @sergeyem.


## Authors

- [@sergeyem](https://www.sergeyem.ru)
