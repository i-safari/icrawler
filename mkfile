TARG=iwatcher

all:Q:
  go get -u -v github.com/themester/icrawler
  go build -o $TARG
