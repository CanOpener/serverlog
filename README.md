Server Log v1.2.0
=================
A thread safe logging system for servers. Serverlog can log to files and to the console(with pretty colours) depending on what you specify. Logfiles will be separated by dates and can also manage how many logfiles you want on the server at any one time.

## Import
Go get the package like this
```
go get github.com/canopener/serverlog
```
Import like this
```
import "github.com/canopener/serverlog"
```

## Usage
The package needs to be initialized. Use the Init function for this.

1. Set the first parameter to true to enable logging to the console.

2. Set the second Parameter to true to enable logging to a file.

3. The third parameter represents the maximum amount of logfiles you wish to have at any one time. the logfiles are separated by days e.g "20-12-2015.csrv.log", "21-12-2015.csrv.log" etc.. Set to anything less than 1 if you want there to be no limit. The oldest logfiles will be deleted as new ones are created in order have only the amount of logfiles you specify here in the directory at any one time.

4. The fourth parameter is the path to the directory in wich you wish to store the logfiles. You need to have read and write permissions to the directory in order for Serverlog to work. Serverlog needs to delete old logfiles to make sure that the third parameter is satisfied so make sure you dont put anything of value within this directory as it has a chance of being deleted

Note: if the second parameter is set to false the third and fourth can be random as they wont even be evaluated by serverlog.

```
serverlog.Init(true, true, 7, "/home/joe/logs/myServerLogs")
```

4 logging functions are available.
```
serverlog.Startup("server listening on port:", PORT) // for startup logging

serverlog.General("Accepted connection from:", conn.IP) // for general logging

serverlog.Warning(conn.IP, "sending lots of data, possibly DOS attack?")

serverlog.Fatal(conn.IP, "crashed the server! Kill all") // This will terminate the program with an exit code of 1
```

![Console Demo](http://i.imgur.com/jYQHbMc.png)

![File Demo](http://i.imgur.com/SPWENaK.png?1)

The log listener can be terminated with :
```
serverlog.Kill()
```
When this happens the serverlog package needs to be initialized before it can log again.

## License
MIT License

Use however you want. A reference would be nice but not mandatory.
