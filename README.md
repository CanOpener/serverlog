# serverlog
A thread safe logging system wit colours for servers. Thread safe for writing the log to a file

## Import
Go get the package like this
```
go get https://github.com/canopener/serverlog
```
Import like this
```
import "github.com/canopener/serverlog"
```

## Usage
The package needs to be initialised. Use the init function for this.

Set the first parameter to true to enable logging to the console.

Set the second Parameter to true to enable logging to a file.

If the second parameter is set to true the third parameter needs to be the path to the logfile(doesnt have to exist, if exists will append)

```
serverlog.Init(true, true, "/home/mladen/logs/myserver.log")
```

3 logging functions are available.
```
serverlog.Startup("server listening on port:", PORT) // for startup logging

serverlog.General("Accepted connection from:", conn.IP) // for general logging

serverlog.Warning(conn.IP, "sending lots of data, possibly DOS attack?")

serverlog.Fatal(conn.IP, "crashed the server! Kill all") // This will terminate the program with an exit code of 1
```

![Console Demo](http://i.imgur.com/jYQHbMc.png)

![File Demo](http://i.imgur.com/SPWENaK.png?1)

## License
MIT License

Use however you want. A reference would be nice but not mandatory.
