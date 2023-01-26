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

## Command line arguments

You can run executable script with following arguments:

* `dsn` - MySQL data source name. Pass this param if you want to rewrite ENV variable.
* `db-max-open-conns` - MySql max open connections. Default is 25.
* `db-max-idle-conns` - Mysql maximum idle connections. Default is 25.
* `db-max-idle-time` - MySql max connection idle time. Default 15m.
* `env` - Environment (development,staging or production)
* `registration` - Is registration enabled? Default true.
* `confirmation` - Is email confirmation enabled? Default true.
* `smtp-host` - SMTP Host
* `smtp-port` - SMTP port
* `smtp-username` - SMTP Username
* `smtp-password` - SMTP Password
* `smtp-sender` - SMTP Sender
* `domain` - What is domain name of server? Default is http://easylist.sergeyem.ru
* `limiter-rps` - Rate limiter maximum requests per second. Default is 2
* `limiter-burst` - Rate limiter maximum burst. Default is 4.
* `limiter-enabled` - Enable rate limiter or not? Default is true.
* `cors-trusted-origins` - Trusted cors origins (space separated)
* `version` - Display script version and exit.

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
