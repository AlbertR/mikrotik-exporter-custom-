# Mikrotik metrics exporter (custom services)
This is a fork of the repository https://github.com/nshttpd/mikrotik-exporter adjusted for monitoring additional metrics.

Observed Services:

> System -> Resource
>
> Metrics:
```
free-memory
total-memory
cpu-load
free-hdd-space
total-hdd-space
uptime
board-name
version
```

> Firmware
>
> Metrics:
```
board-name
model
serial-number
current-firmware
upgrade-firmware
```

> Interfaces
>
> Metrics:
```
name
comment
rx-byte
tx-byte
rx-packet
tx-packet
rx-error
tx-error
rx-drop
tx-drop
```

> Wireles Interface and Status
>
> Metrics:
```
interface
mac-address
signal-to-noise
signal-strength-ch0
packets
bytes
frames
channel
registered-clients
noise-floor
overall-tx-ccq
```

> DHCP Servers and Leases
>
> Metrics:
```
name
address
server
active-mac-address
status
expires-after
active-address
host-name
```

>DHCP Pool
>
> Metrics:
```
name
next-pool
ranges
```

> IP Addresses
>
> Metrics:
```
name
address
interface
netmask
network
```
