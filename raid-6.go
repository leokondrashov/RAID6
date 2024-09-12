package main

import (
    "fmt"
    "strconv"
	"os"
)

type File struct {
    name string
    offset int
    blockCount int
    size int
}

type Disk struct {
    blockCount  int
    path string
    isWorking bool
}

type Raid6 struct {
    disks [16]Disk
    diskNumber int
    blockSize int
    parityBlockCount int
    files map[string]File
}

func NewRaid6(diskNumber int, blockSize int) *Raid6 {
    var disks [16]Disk
    var files map[string]File = make(map[string]File)
    for i := 0; i < diskNumber; i++ {
        path := "./disk"+strconv.Itoa(i)
        data := path + "/data"
        os.Mkdir(path, os.ModePerm)
        os.Create(data)
        var disk Disk
        disk.blockCount = 0
        disk.path = path
        disk.isWorking = true
        disks[i] = disk
    }
    os.Create("parityP")
    os.Create("parityQ")

    fmt.Println("raid-6 is created\n")
    
	return &Raid6{
		disks:            disks,
        diskNumber:       diskNumber,
		blockSize:        blockSize,
		parityBlockCount: diskNumber - 2,
        files:            files,
	}
}

func (raid6 *Raid6) writeData(filename string, data []byte) error {
    size := len(data)
    blockCount := (size + raid6.blockSize - 1) / raid6.blockSize
    blockCountPerDisk := (blockCount + raid6.diskNumber - 1) / raid6.diskNumber
    diskSize := blockCountPerDisk * raid6.diskNumber * raid6.blockSize

    fmt.Println("size: ", size)
    fmt.Println("diskSize: ", diskSize)
    for i := size; i < diskSize; i++ {
        data = append(data, 0)
    }
    
    chunks := make([][]byte, blockCountPerDisk * raid6.diskNumber)
    for i := 0; i < blockCountPerDisk * raid6.diskNumber; i++ {
        chunks[i] = make([]byte, raid6.blockSize)
        for j := 0; j < raid6.blockSize; j++ {
            chunks[i][j] = data[i * raid6.blockSize + j]
        }
    }
    // fmt.Println(len(chunks))
    // fmt.Println(len(chunks[0]))
    raid6.calculateParity(chunks)
    
    var file File
    file.name = filename
    file.size = size
    file.offset = raid6.disks[0].blockCount
    file.blockCount = blockCountPerDisk
    raid6.files[filename] = file
    for i := 0; i < raid6.diskNumber; i++ {
        raid6.disks[i].blockCount += blockCountPerDisk
    }

    diskCursor := 0
    
    for i := 0; i < blockCountPerDisk * raid6.diskNumber; i++ {
        if diskCursor >= raid6.diskNumber {
            diskCursor -= raid6.diskNumber
        }
        diskPath := "./disk" + strconv.Itoa(diskCursor) + "/data"
        f, _ := os.OpenFile(diskPath, os.O_APPEND|os.O_WRONLY, 0644)
        f.Write(data[i * raid6.blockSize : (i+1) * raid6.blockSize])
    }

    return nil
}

func (raid6 *Raid6) calculateParity(chunks [][]byte) {

}

func (raid6 *Raid6) readData(filename string) []byte {
    file := raid6.files[filename]
    offset := file.offset
    blockCountPerDisk := file.blockCount

    var result []byte
    buf := make([]byte, raid6.blockSize)

    diskCursor := 0
    size := 0
    var offsetLocal int64
	for i := 0; i < blockCountPerDisk * raid6.diskNumber; i++ {
        offsetLocal = int64(offset + i * raid6.blockSize)
        if diskCursor >= raid6.diskNumber {
            diskCursor -= raid6.diskNumber
        }
        diskPath := "./disk" + strconv.Itoa(diskCursor) + "/data"
        f, _ := os.Open(diskPath)
        n, _ := f.ReadAt(buf, offsetLocal)
        fmt.Println("filesize: ", file.size, " sizenow: ", size)
        if size + n <= file.size {
            result = append(result, buf[:n]...)
            size += raid6.blockSize
        } else {
            result = append(result, buf[:file.size - size]...)
            break
        }
	}
	return result
}



func main() {
	fmt.Println("test begins\n")
	r := NewRaid6(4, 20)

	var data []byte
    for i := 0; i < 5 * 20 + 16; i++ {
        data = append(data, byte(i % 20))
    }

    fmt.Println(data)
    
    r.writeData("test", data)

    res := r.readData("test")
    for _, value := range res {
        fmt.Println(value, " ")
    }
}
