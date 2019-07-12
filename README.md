# find-failing-notifications

This is a utility that scans Vidispine logs that have been warehoused in Elasticsearch (usually via Logstash)
and looks for NETWORK_FAILURE errors relating to notifications.

It compiles a list of each failing notification and then queries the server for
information about them in order to make debugging easier.

## How to use it

We use the utility from within a Docker image so it gets settings from the environment.
You should set the following environment variables:

- *ELASTICSEARCH_URL* - URL to get the elasticsearch logs from. Defaults to http://localhost:9200.
- *VIDISPINE_URI* - URL to access the Vidispine API. Don't include the /API prefix. No default. Example: http://localhost:8080
- *VIDISPINE_USER* - username to access the Vidsipine API
- *VIDISPINE_PASSWD* - password to access the Vidispine API
- *INDEX_NAME* - index name (or pattern) on the Elasticsearch server that contains the Vidispine logs.

With all of these parameters set, simply run the app.  It will scan for NETWORK_FAILURE logs and if they contain
a reference to a notification then information about this notification will be shown.

## Log record format

It's expected that there is a field called `message_detail` in the log record and this is what is scanned.
The (subset) of expected data is shown in the `Record` struct in `models.go`. If your index stores data
differently you'll have to update this.

## How to build
You'll need Go v1.11 or higher installed (or run in a Docker container) in order to compile the software.

If you're not used to Go, remember:
- it compiles to statically linked native code. Just copy the binary and run - no need to set
up a runtime environment or install libraries
- so you only need Go installed to compile, not to run
- it automatically builds for your environment, be it Linux, Windows, Mac, AIX, Solaris, Hurd, etc. etc.
Simply check out the repo and run:

```
$ go test
PASS
ok  	github.com/fredex42/find-failing-notifications	0.007s
$ go build
$ (set up environment variables)
$ ./find-failing-notifications
```
