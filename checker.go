package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/sys/unix"
)

const (
	size2Mi = 2 * (1 << (2 * 10))
	size1Gi = 1 * (1 << (3 * 10))
)

type HugePageMount struct {
	Size int
	Path string
}

func log(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
func log_info(format string, args ...interface{}) {
	log("INFO: "+format, args...)
}
func log_warn(format string, args ...interface{}) {
	log("WARNING: "+format, args...)
}
func log_error(format string, args ...interface{}) {
	log("ERROR: "+format, args...)
}

func getHugePagesMounts() ([]HugePageMount, error) {
	log_info("Opening /proc/mounts")
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open /proc/mounts file")
	}
	defer f.Close()

	r := bufio.NewReader(f)

	mounts := make([]HugePageMount, 0)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, errors.Wrapf(err, "Unable to read /proc/mounts file")
			}
		}

		mountArguments := strings.Split(line, " ")
		if len(mountArguments) < 4 {
			continue
		}
		if mountArguments[2] != "hugetlbfs" {
			continue
		}

		pageSize := 0
		mountParams := strings.Split(mountArguments[3], ",")
		for _, param := range mountParams {
			if !strings.HasPrefix(param, "pagesize=") {
				continue
			}
			if pageSize != 0 {
				log_warn("pagesize of mount already scanned '%s'", line)
				continue
			}
			var size int
			var unit string
			if _, err := fmt.Sscanf(param, "pagesize=%d%s", &size, &unit); err != nil {
				log_error("scanning pagesize of mount '%s':%v", line, err)
				continue
			}
			switch unit {
			case "M":
				pageSize = size << (10 * 2)
			case "G":
				pageSize = size << (10 * 3)
			default:
				log_error("unknown unit of pagesize of mount '%s'", line)
				continue
			}
		}
		if pageSize == 0 {
			log_warn("wasn't able to detect pagesize of mount '%s'", line)
			continue
		}

		mounts = append(mounts, HugePageMount{
			Path: mountArguments[1],
			Size: pageSize,
		})
	}

	return mounts, nil
}

func (hpm *HugePageMount) testPage() error {
	log_info("Testing %s", hpm.Path)
	fileName := filepath.Join(hpm.Path, fmt.Sprintf("checker_%d", time.Now().Unix()))

	log_info("Creating file %s", fileName)
	fd, err := unix.Open(fileName, unix.O_RDWR|unix.O_CREAT, 0755)
	if err != nil {
		return errors.Wrapf(err, "Create hugepage file syscall error")
	}
	defer unix.Unlink(fileName)

	log_info("Trying to mmap 1 page of %v size", hpm.Size)
	page, err := unix.Mmap(fd, 0, hpm.Size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return errors.Wrapf(err, "Unable to map page")
	}
	defer unix.Munmap(page)

	log_info("Trying to write page data")
	for i := range page {
		page[i] = byte(i)
	}

	valid := true
	log_info("Trying to read page data and validate it")
	for i := range page {
		if page[i] != byte(i) {
			valid = false
		}
	}

	if !valid {
		return fmt.Errorf("Data corrupted on validation stage")
	}
	log_info("Test passed for %s", hpm.Path)
	return nil
}

func testHugepages() error {
	mounts, err := getHugePagesMounts()
	if err != nil {
		return errors.Wrap(err, "Error when get list of pages mount")
	}

	log_info("Mounts found: %+v", mounts)
	for _, mount := range mounts {
		if err = mount.testPage(); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := testHugepages(); err != nil {
		log_error("%v", err)
		os.Exit(1)
	}
}
