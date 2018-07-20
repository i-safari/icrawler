# icrawler
Instagram (night)crawler.

## Installation
```
go get -u -v github.com/dgrr/icrawler
go install github.com/dgrr/icrawler

# Add $GOPATH/bin to $PATH env var in bashrc
# Uncomment the following line.
# echo 'PATH=$PATH:$GOPATH/bin/ export PATH' >> ~/.bashrc
```

## Update

You can just use `go get -u -v github.com/dgrr/icrawler`.

## Usage

After compile it you can start it using command parameters.

```bash
$ icrawler -h
Usage of icrawler:
  -d string
    	Instagram database (default "./instagram.db")
  -g int
    	Telegram chat id
  -l string
    	Log file (default "./icrawler.log")
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

Any target of file can be followed by:
- 'f': Does not store and send followers notification.
- 'w': Does not store and send following notification.
- 'm': Does not store and send media.
- 's': Does not store and send stories.
- 'p': Does not store and send profile changes.
- 'h': Does not store and send highlights.
- 'n': Only store and send new media.

```
elonmusk f w  # do not store followers and following
```
