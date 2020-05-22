# LOGGING


## Defaults for logging

Use of the standard Go `log` package is deprecated and should be avoided. 
The recommended way of logging is through the logger `util.Logger`, which supports [all the functions normally available to `log`](https://golang.org/pkg/log/#Logger).


By default, **logging is disabled**. Any `Logger.Printf` statement will simply be discarded.

To enable logging, you should use

```go
util.EnableLogging = true
util.SetLog()
```

When enabled, the default output for logging is a file named `go-vcloud-director.log`.
The file name can be changed using

```go
util.ApiLogFileName = "my_file_name.log"
```


If you want logging on the screen, use

```go
util.Logger.SetOutput(os.Stdout)
```

or

```
util.Logger.SetOutput(os.Stderr)
```

## Automatic logging of HTTP requests and responses.

The HTTP requests and responses are automatically logged.
Since all the HTTP operations go through `NewRequest` and `decodeBody`, the logging captures the input and output of the request with calls to `util.ProcessRequestOutput` and `util.ProcessResponseOutput`.

These two functions will show the request or response, and the function from which they were called, giving devs an useful tracking tool.

The output of these functions can be quite large. If you want to mute the HTTP processing, you can use:

```go
util.LogHttpRequest = false
util.LogHttpResponse = false
```

During the request and response processing, any password or authentication token found through pattern matching will be automatically hidden. To show passwords in your logs, use

```go
util.LogPasswords = true
```

It is also possible to skip the output of the some tags (such as the result of `/versions` request,) which are quite large using 

```go
util.SetSkipTags("SupportedVersions,ovf:License")
```

For an even more dedicated log, you can define from which function names you want the logs, using

```go
util.SetApiLogFunctions("FindVAppByName,GetAdminOrgByName")
```

## Custom logger

If the configuration options are not enough for your needs, you can supply your own logger.

```go
util.SetCustomLogger(mylogger)
```

## Environment variables

The logging behavior can be changed without coding. There are a few environment variables that are checked when the library is used:

Variable                    | Corresponding environment var 
--------------------------- | :-------------------------------
`EnableLogging`             | `GOVCD_LOG`
`ApiLogFileName`            | `GOVCD_LOG_FILE`
`LogPasswords`              | `GOVCD_LOG_PASSWORDS`
`LogOnScreen`               | `GOVCD_LOG_ON_SCREEN`
`LogHttpRequest`            | `GOVCD_LOG_SKIP_HTTP_REQ`
`LogHttpResponse`           | `GOVCD_LOG_SKIP_HTTP_RESP`
`SetSkipTags`               | `GOVCD_LOG_SKIP_TAGS`
`SetApiLogFunctions`        | `GOVCD_LOG_INCLUDE_FUNCTIONS`
`OverwriteLog`              | `GOVCD_LOG_OVERWRITE`
