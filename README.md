# IPv4 Address Counter

Calculates the number of __unique addresses__ in a file.
Input file is a simple text file with IPv4 addresses. One line is one address.

The file is unlimited in size and can occupy tens and hundreds of gigabytes.

## Implementation
IPv4 uses a 32-bit address space which provides 4294967296 (2^32) unique addresses.

So we need 4294967296 bits to track unique IPs. It's 4294967296 / 8 = 536,870,912 bytes,
or 256 X 2,097,152 Matrix.

The program reads the file by byte chunks and send them to a CHANNEL.
Workers read the chunks and process them, checking the state Matrix and updating the counter.

Each Matrix column (256) has a mutex (lock) to sync threads only there and improve concurrent processing speed.

## Run
```
.\ip-addr-counter.exe -f="FILE_PATH"
OR
.\ip-addr-counter.exe   //expect file ip_addresses in the same folder
```

## Tests
Results on the provided file (~120 GB):

**Number of unique IPv4:** 1000000000

**Time:** 1m17.555218s

**CPU:** AMD Ryzen 9 5900X 12-Core

```
GO ENV
set GO111MODULE=on
set GOARCH=amd64
set GOHOSTARCH=amd64
set GOHOSTOS=windows
set GOVERSION=go1.22.7
set GCCGO=gccgo
set GOAMD64=v1
set AR=ar
set CC=gcc
set CXX=g++
set CGO_ENABLED=0
```

```
(pprof) top
Showing nodes accounting for 620.78MB, 99.42% of 624.37MB total
Dropped 6 nodes (cum <= 3.12MB)
      flat  flat%   sum%        cum   cum%
  600.75MB 96.22% 96.22%   602.66MB 96.52%  main.getUniqueIpv4AddrNumber
   10.24MB  1.64% 97.86%    20.02MB  3.21%  main.worker
    9.78MB  1.57% 99.42%     9.78MB  1.57%  strings.genSplit
         0     0% 99.42%   603.82MB 96.71%  main.main
         0     0% 99.42%   604.35MB 96.79%  runtime.main
         0     0% 99.42%     9.78MB  1.57%  strings.Split (inline)

```
