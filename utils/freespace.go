package utils

import (
	"math"
	"os"
	"strconv"
	"syscall"
)

type Disk struct {
	Total uint64
	Used  uint64
	Free  uint64
}

const (
	b  = 1
	kb = 1024 * b
	mb = 1024 * kb
	gb = 1024 * mb
)

// FreeSpaceHome free space in home directory
func FreeSpaceHome() (d *Disk) {
	home, err := homeDir()
	d = CheckFreeSpace(home)
	if err != nil {
		return
	}

	return
}

// CheckFreeSpace free space of the specified directory
func CheckFreeSpace(path string) (d *Disk) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}

	d = &Disk{
		Total: fs.Blocks * uint64(fs.Bsize),
		Free:  fs.Bfree * uint64(fs.Bsize),
		Used:  (fs.Blocks * uint64(fs.Bsize)) - (fs.Bfree * uint64(fs.Bsize)),
	}
	return
}

// GetPercentFree free space percentage
func (d Disk) GetPercentFree() float64 {
	return math.Round(100.0 * float64(d.Free) / float64(d.Total))
}

// GetFreeSpace free space
func (d Disk) GetFreeSpace() float64 {
	return float64(d.Free)
}

// CalculateSpaceBeforeDeploy calculate free space before deployment
func (d Disk) CalculateSpaceBeforeDeploy(deployFilesSize float64) float64 {
	return math.Round(float64(d.Free) - deployFilesSize)
}

// ToAvailablePercent calculate free percent
func ToAvailablePercent(size uint64, total uint64) float64 {
	return math.Round(100.0 * float64(size) / float64(total))
}

// HumanSize convert bytes to human friendly format
func HumanSize(size float64) string {
	var suffixes [5]string

	suffixes[0] = "B"
	suffixes[1] = "KB"
	suffixes[2] = "MB"
	suffixes[3] = "GB"
	suffixes[4] = "TB"

	base := math.Log(size) / math.Log(1024)
	getSize := round(math.Pow(1024, base-math.Floor(base)), .5, 1)
	getSuffix := suffixes[int(math.Floor(base))]
	return strconv.FormatFloat(getSize, 'f', -1, 64) + " " + string(getSuffix)
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func homeDir() (string, error) {
	return os.UserHomeDir()
}
