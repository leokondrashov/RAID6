# RAID6
Project for CE7490 Distributed Systems

The project report is available at this Overleaf link: https://www.overleaf.com/read/skfnpzfqhxxt#2f02d1.

## Description

This project implements RAID6 with the following features:

* Arbitrary disk configuration for data and checksum shards
* Arbitrary-sized files store and read
* Recovery from disk failures
* Simple filesystem: store and read by file name

## Usage

### Command-line interface

We use go >1.21 for our implementation. Please install the appropriate version before running.

Each RAID operation is a separate execution of main.go with specific flags to choose the operation to perform and the parameters of the operation:

```
Usage:
go run main.go [options] COMMAND [parameters]

COMMANDS:
  store [file]
        Stores the file into RAID
  read [file] [dstFile]
        Reads file from RAID and writes it into dstFile
  recover
        Recovers from disk failure

Options of main.go:
  -classic
        Use classic RAID6 Linux implementation
  -data int
        Number of data disks (default 6)
  -dir string
        Directory to use for the shards (default "data")
  -parity int
        Number of parity disks (default 2)
  -raid string
        RAID filesystem records file (default "raid.json")
```

Shards are stored as files in `data` directory. We simulate disk failure as the deletion of some of the files.

You can change the RAID configuration by passing `-data` and `-parity` flags to each operation.

### Example scenario

```
# store file test.txt
go run main.go store test.txt

# read stored file test.txt into test2.txt
go run main.go read test.txt test2.txt

# simulate failure in disk 1
rm data/shard1
# recover the missing shards
go run main.go recover

# read stored file test.txt into test3.txt
go run main.go read test.txt test3.txt
```


