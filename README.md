# icrawler
Instagram (night)crawler.

## Installation
```
go get -u -v github.com/themester/icrawler
```

## Update

You can just use `go get -u` or execute `mk` at $GOPATH/src/github.com/themester/icrawler.

[Download mk](https://github.com/9fans/plan9port)

## Usage

After compile it you can start it using command parameters.

```bash
$ /icrawler -h
Usage of ./icrawler:
  -d string
    	Instagram database (default "./instagram.db")
  -g int
    	Telegram chat id
  -l string
    	Log file (default "./iwatcher.log")
  -n string
    	Telegram bot api id
  -o string
    	Output directory or storing directory (default "./files")
  -t string
    	Targets file (default "./targets")
  -u uint
    	Update time in minutes (default 5)
```

You must use `-u` and `-n` parameters. To know your ID use @userinfobot

```bash
$ icrawler -n "{You bot API key}" -g {Your user id} -u 5 & disown
```

Bot can be stopped using SIGINT signal. Try this:

```bash
$ pkill -2 icrawler
```

Targets file can be modified during bot running. You can just add new user executing

```bash
$ echo 'my_girlfriend/boyfriend_profile' >> targets
```
