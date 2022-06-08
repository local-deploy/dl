package helper

import (
	"math"
	"syscall"
)

type disk struct {
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
func FreeSpaceHome() (d *disk) {
	home, err := HomeDir()
	d = CheckFreeSpace(home)
	if err != nil {
		return
	}

	return
}

// CheckFreeSpace free space of the specified directory
func CheckFreeSpace(path string) (d *disk) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}

	d = &disk{
		Total: fs.Blocks * uint64(fs.Bsize),
		Free:  fs.Bfree * uint64(fs.Bsize),
		Used:  (fs.Blocks * uint64(fs.Bsize)) - (fs.Bfree * uint64(fs.Bsize)),
	}
	return
}

// GetPercentFree free space percentage
func (d disk) GetPercentFree() float64 {
	return math.Round(100.0 * float64(d.Free) / float64(d.Total))
}

// GetFreeSpace free space
func (d disk) GetFreeSpace() float64 {
	return float64(d.Free)
}

// CalculateSpaceBeforeDeploy calculate free space before deployment
func (d disk) CalculateSpaceBeforeDeploy(deployFilesSize float64) float64 {
	return math.Round(float64(d.Free) - deployFilesSize)
}

// ToKb convert bytes to kilobytes
func ToKb(bytes float64) float64 {
	return math.Round(bytes / float64(kb))
}

// ToMb convert bytes to megabytes
func ToMb(bytes float64) float64 {
	return math.Round(bytes / float64(mb))
}

// ToGb convert bytes to gigabytes
func ToGb(bytes float64) float64 {
	return math.Round(bytes / float64(gb))
}

// ToAvailablePercent calculate free percent
func ToAvailablePercent(size uint64, total uint64) float64 {
	return math.Round(100.0 * float64(size) / float64(total))
}
