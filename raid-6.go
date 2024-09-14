package main

import (
	"fmt"
	"os"
	"strconv"
)

type File struct {
	name       string
	offset     int
	blockCount int
	size       int
}

type Disk struct {
	blockCount int
	path       string
	isWorking  bool
}

type Raid6 struct {
	disks      [16]Disk
	diskNumber int
	blockSize  int
	files      map[string]File
}

func NewRaid6(diskNumber int, blockSize int) *Raid6 {
	var disks [16]Disk
	var files map[string]File = make(map[string]File)
	for i := 0; i < diskNumber; i++ {
		path := "./disk" + strconv.Itoa(i)
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
		disks:      disks,
		diskNumber: diskNumber,
		blockSize:  blockSize,
		files:      files,
	}
}

const (
	fieldSize = 256
	generator = 0x1d // Irreducible polynomial used in GF(2^8)
)

var (
	expTable [fieldSize]byte
	logTable [fieldSize]byte
)

func init() {
	// Initialize the exponentiation (expTable) and logarithm (logTable) tables for GF(2^8)
	var x byte = 1
	for i := 0; i < fieldSize; i++ {
		expTable[i] = x
		logTable[x] = byte(i)
		x = galoisFieldMulInternal(x, 0x02) // Multiply by the generator (0x02)
	}
}

// Galois field exponentiation using the precomputed table
func galoisFieldExp(base byte, exponent byte) byte {
	if base == 0 {
		return 0
	}
	logBase := logTable[base]
	return expTable[(int(logBase)*int(exponent))%(fieldSize-1)]
}

// Galois field multiplication
func galoisFieldMul(a byte, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	logA := logTable[a]
	logB := logTable[b]
	return expTable[(int(logA)+int(logB))%(fieldSize-1)]
}

// Galois field inversion
func galoisFieldInv(a byte) byte {
	if a == 0 {
		return 0 // No inverse for zero
	}
	logA := logTable[a]
	return expTable[(fieldSize-1)-int(logA)]
}

// Galois field division
func galoisFieldDiv(a byte, b byte) byte {
	if a == 0 {
		return 0
	}
	if b == 0 {
		panic("Division by zero")
	}
	logA := logTable[a]
	logB := logTable[b]
	return expTable[(int(logA)-int(logB)+(fieldSize-1))%(fieldSize-1)]
}

// Helper function for internal multiplication
func galoisFieldMulInternal(a byte, b byte) byte {
	var result byte = 0
	for b > 0 {
		if (b & 1) != 0 {
			result ^= a
		}
		if (a & 0x80) != 0 {
			a = (a << 1) ^ generator // Reduce modulo the generator polynomial
		} else {
			a <<= 1
		}
		b >>= 1
	}
	return result
}

func (raid6 *Raid6) writeData(filename string, data []byte) error {
	size := len(data)
	blockCount := (size + raid6.blockSize - 1) / raid6.blockSize
	blockCountPerDisk := (blockCount + raid6.diskNumber - 1) / raid6.diskNumber
	diskSize := blockCountPerDisk * raid6.diskNumber * raid6.blockSize

	for i := size; i < diskSize; i++ {
		data = append(data, 0)
	}

	chunks := make([][]byte, blockCountPerDisk*raid6.diskNumber)
	for i := 0; i < blockCountPerDisk*raid6.diskNumber; i++ {
		chunks[i] = make([]byte, raid6.blockSize)
		for j := 0; j < raid6.blockSize; j++ {
			chunks[i][j] = data[i*raid6.blockSize+j]
		}
	}

	raid6.calculateParity(chunks, blockCountPerDisk, raid6.disks[0].blockCount)

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

	for i := 0; i < blockCountPerDisk*raid6.diskNumber; i++ {
		if diskCursor >= raid6.diskNumber {
			diskCursor -= raid6.diskNumber
		}
		diskPath := "./disk" + strconv.Itoa(diskCursor) + "/data"
		f, _ := os.OpenFile(diskPath, os.O_APPEND|os.O_WRONLY, 0644)
		f.Write(data[i*raid6.blockSize : (i+1)*raid6.blockSize])
		diskCursor++
	}

	return nil
}

func (raid6 *Raid6) calculateParity(chunks [][]byte, blockCountPerDisk int, offset int) {
	parityP := make([]byte, blockCountPerDisk*raid6.blockSize)
	parityQ := make([]byte, blockCountPerDisk*raid6.blockSize)

	for i := 0; i < blockCountPerDisk; i++ {
		for j := 0; j < raid6.blockSize; j++ {
			for h := 0; h < raid6.diskNumber; h++ {
				parityP[i*raid6.blockSize+j] ^= chunks[raid6.diskNumber*i+h][j]
				parityQ[i*raid6.blockSize+j] ^= galoisFieldMul(chunks[raid6.diskNumber*i+h][j], expTable[h])
			}
		}
	}

	p, _ := os.OpenFile("./parityP", os.O_APPEND|os.O_WRONLY, 0644)
	q, _ := os.OpenFile("./parityQ", os.O_APPEND|os.O_WRONLY, 0644)
	p.Write(parityP[:])
	q.Write(parityQ[:])
}

func (raid6 *Raid6) recoveryOneDiskWithP(diskId int) {
	newdata, _ := os.ReadFile("./parityP")
	for i := 0; i < raid6.diskNumber; i++ {
		if i == diskId {
			continue
		}
		diskPath := "./disk" + strconv.Itoa(i) + "/data"
		data, _ := os.ReadFile(diskPath)
		for j := 0; j < raid6.blockSize*raid6.disks[0].blockCount; j++ {
			newdata[j] ^= data[j]
		}
	}
	diskPath := "./disk" + strconv.Itoa(diskId) + "/data"
	f, _ := os.OpenFile(diskPath, os.O_WRONLY, 0644)
	f.WriteAt(newdata[:], 0)
}

func (raid6 *Raid6) recoveryOneDiskWithQ(diskId int) {
	newdata, _ := os.ReadFile("./parityQ")
	for i := 0; i < raid6.diskNumber; i++ {
		if i == diskId {
			continue
		}
		diskPath := "./disk" + strconv.Itoa(i) + "/data"
		data, _ := os.ReadFile(diskPath)
		for j := 0; j < raid6.blockSize*raid6.disks[0].blockCount; j++ {
			power := galoisFieldExp(2, byte(i))
			newdata[j] ^= galoisFieldMul(power, data[j])
		}
	}

	for i := 0; i < raid6.blockSize*raid6.disks[0].blockCount; i++ {
		gx := galoisFieldExp(2, byte(diskId))
		newdata[i] = galoisFieldDiv(newdata[i], gx)
	}

	diskPath := "./disk" + strconv.Itoa(diskId) + "/data"
	f, _ := os.OpenFile(diskPath, os.O_WRONLY, 0644)
	f.WriteAt(newdata[:], 0)
}

func (raid6 *Raid6) recoveryTwoDisk(diskId1 int, diskId2 int) {
	parityP, _ := os.ReadFile("./parityP")
	parityQ, _ := os.ReadFile("./parityQ")

	for i := 0; i < raid6.diskNumber; i++ {
		if i == diskId1 || i == diskId2 {
			continue
		}
		diskPath := "./disk" + strconv.Itoa(i) + "/data"
		data, _ := os.ReadFile(diskPath)
		for j := 0; j < raid6.blockSize*raid6.disks[0].blockCount; j++ {
			parityP[j] ^= data[j]
			power := galoisFieldExp(2, byte(i))
			parityQ[j] ^= galoisFieldMul(data[j], power)
		}
	}

	recovered1 := make([]byte, raid6.blockSize*raid6.disks[0].blockCount)
	recovered2 := make([]byte, raid6.blockSize*raid6.disks[0].blockCount)

	gx_inv := galoisFieldInv(galoisFieldExp(2, byte(diskId1)))
	gyx := galoisFieldMul(galoisFieldExp(2, byte(diskId2)), gx_inv)
	gyx_1 := gyx ^ 1

	for i := 0; i < raid6.blockSize*raid6.disks[0].blockCount; i++ {
		recovered1[i] = galoisFieldDiv(
			galoisFieldMul(gx_inv, parityQ[i])^galoisFieldMul(gyx, parityP[i]),
			gyx_1,
		)
		recovered2[i] = parityP[i] ^ recovered1[i]
	}

	diskPath := "./disk" + strconv.Itoa(diskId1) + "/data"
	f, _ := os.OpenFile(diskPath, os.O_WRONLY, 0644)
	f.WriteAt(recovered1[:], 0)

	diskPath = "./disk" + strconv.Itoa(diskId2) + "/data"
	g, _ := os.OpenFile(diskPath, os.O_WRONLY, 0644)
	g.WriteAt(recovered2[:], 0)
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
	for i := 0; i < blockCountPerDisk*raid6.diskNumber; i++ {
		offsetLocal = int64(offset + (i/raid6.diskNumber)*raid6.blockSize)
		if diskCursor >= raid6.diskNumber {
			diskCursor -= raid6.diskNumber
		}
		diskPath := "./disk" + strconv.Itoa(diskCursor) + "/data"
		f, _ := os.Open(diskPath)
		n, _ := f.ReadAt(buf, offsetLocal)
		if size+n <= file.size {
			result = append(result, buf[:n]...)
			size += raid6.blockSize
			diskCursor++
		} else {
			result = append(result, buf[:file.size-size]...)
			break
		}
	}
	return result
}

func main() {
	fmt.Println("test begins\n")
	r := NewRaid6(4, 20)

	var data []byte
	var invalid []byte
	for i := 0; i < 5*20+60; i++ {
		data = append(data, byte(i%20))
		invalid = append(invalid, 0)
	}

	fmt.Println("origin data ", data)

	r.writeData("test", data)

	diskPath := "./disk" + strconv.Itoa(1) + "/data"
	f, _ := os.OpenFile(diskPath, os.O_WRONLY, 0644)
	f.WriteAt(invalid[:], 0)

	diskPath = "./disk" + strconv.Itoa(2) + "/data"
	g, _ := os.OpenFile(diskPath, os.O_WRONLY, 0644)
	g.WriteAt(invalid[:], 0)

	inter := r.readData("test")
	fmt.Println("after disk fault ", inter)

	r.recoveryTwoDisk(1, 2)
	// r.recoveryOneDiskWithQ(1)

	res := r.readData("test")
	fmt.Println("after recovery ", res)
}
