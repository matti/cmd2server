# cmd2server


Opens a port and runs a command when client connects outputting the stdout/stderr and terminates the command on client disconnect / command exit before accepting a new one.

Similar to socat exec, but does not leave zombie processes behind and terminates the process before accepting a new connection.

## Usage

Start:

```
$ go run cmd/cmd2server/main.go 0.0.0.0:1234 ping -c 5 google.com
2020/09/19 10:23:39 waiting for connection at 0.0.0.0:1234
```

Then

```
$ nc localhost 1234
PING google.com (172.217.21.174): 56 data bytes
64 bytes from 172.217.21.174: icmp_seq=0 ttl=116 time=24.339 ms
64 bytes from 172.217.21.174: icmp_seq=1 ttl=116 time=23.770 ms
64 bytes from 172.217.21.174: icmp_seq=2 ttl=116 time=17.664 ms
64 bytes from 172.217.21.174: icmp_seq=3 ttl=116 time=25.530 ms
64 bytes from 172.217.21.174: icmp_seq=4 ttl=116 time=25.590 ms

--- google.com ping statistics ---
5 packets transmitted, 5 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 17.664/23.379/25.590/2.941 ms

$
```

and the command logs say:

```
2020/09/19 10:24:17 waiting for connection at 0.0.0.0:1234
2020/09/19 10:24:19 accepted connection from [::1]:53384
2020/09/19 10:24:19 exec ping with args [-c 5 google.com]
PING google.com (172.217.21.174): 56 data bytes
64 bytes from 172.217.21.174: icmp_seq=0 ttl=116 time=24.339 ms
64 bytes from 172.217.21.174: icmp_seq=1 ttl=116 time=23.770 ms
64 bytes from 172.217.21.174: icmp_seq=2 ttl=116 time=17.664 ms
64 bytes from 172.217.21.174: icmp_seq=3 ttl=116 time=25.530 ms
64 bytes from 172.217.21.174: icmp_seq=4 ttl=116 time=25.590 ms

--- google.com ping statistics ---
5 packets transmitted, 5 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 17.664/23.379/25.590/2.941 ms
2020/09/19 10:24:23 read err io: read/write on closed pipe
2020/09/19 10:24:23 killing PID 9456 with signal SIGTERM
2020/09/19 10:24:23 PID 9456 exited with 0
2020/09/19 10:24:23 waiting for connection at 0.0.0.0:1234
```
