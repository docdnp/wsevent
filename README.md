# wsevent

Build and start three dummy services on ports 8080 to 8081

```
go build
./svcdummy -addr localhost:8080 &
./svcdummy -addr localhost:8081 &
./svcdummy -addr localhost:8082 &
```

Start the loadbalancer and reverse proxy
```
./wsproxy.sh
```

Start the multiple clients using the loadbalancer

```
./wsproxy.sh 80
```
