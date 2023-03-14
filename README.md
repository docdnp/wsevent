# wsevent

## Scenario: Client consumes over reverse proxy
Build and start three dummy services on ports 8080 to 8081

```
go build
./svcdummy -serve-address localhost:8080 &
./svcdummy -serve-address localhost:8081 &
./svcdummy -serve-address localhost:8082 &
```

or 

```
go run svcdummy -serve-address localhost:8080 &
go run svcdummy -serve-address localhost:8081 &
go run svcdummy -serve-address localhost:8082 &
```

Start the loadbalancer and reverse proxy
```
./wsproxy.sh
```

Start multiple clients using the loadbalancer

```
./wsclient.sh 80
```

## Scenario: Producer -> ProxyCLient -> ProxyCLient
Start one producing service instance that writes random strings on `ws://localhost:8080/echo`:

```
go run svcdummy &
```

Start one service instance reading from `ws://localhost:8080/echo` and proxying to `ws://localhost:8888/proxy-echo`:

```
go run svcdummy -consume-path echo -consume-address localhost:8080 -serve-address localhost:8888 -serve-path proxy-echo -no-produce &
```

Start a second service instance reading from `ws://localhost:8888/proxy-echo` and proxying to `ws://localhost:9999/the-end`:

```
go run svcdummy -consume-path proxy-echo -consume-address localhost:8888 -serve-address localhost:9999 -serve-path the-end -no-produce &
```

Start the end consumer:
```
./wsclient.sh 9999
```
