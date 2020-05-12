package main

import (
	"fmt"
	"github.com/shirou/gopsutil/host"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main(){
	fmt.Print("+++++++++++++++++++++++++++++\n")
	// Check system time
	fmt.Println("The system current time is:")
	fmt.Println(osTimeCheck())
	fmt.Print("+++++++++++++++++++++++++++++\n")

	// Check system day
	fmt.Println("The system current day is:")
	fmt.Println(osDayCheck())
	fmt.Print("+++++++++++++++++++++++++++++\n")

	// Check system uptime
	fmt.Println("The system uptime duration is:")
	getUptime()
	fmt.Print("+++++++++++++++++++++++++++++\n")

	// Check disk storage
	fmt.Println("The system disk usage is as follows:")
	disk := DiskUsage("/")
	fmt.Printf("All: %.2f GB\n", float64(disk.All)/float64(GB))
	fmt.Printf("Used: %.2f GB\n", float64(disk.Used)/float64(GB))
	fmt.Printf("Free: %.2f GB\n", float64(disk.Free)/float64(GB))
	fmt.Print("+++++++++++++++++++++++++++++\n")

	//Check memory usage
	fmt.Println("The system memory usage is as follows:")
	getMemUsage()
	var overall [][]int
	for i := 0; i<4; i++ {
		// Allocate memory using make() and append to overall (so it doesn't get
		// garbage collected). This is to create an ever increasing memory usage
		// which we can track. We're just using []int as an example.
		a := make([]int, 0, 999999)
		overall = append(overall, a)
		// Print our memory usage at each interval
		getMemUsage()
		time.Sleep(time.Second)
	}
	// Clear our memory and print usage, unless the GC has run 'Alloc' will remain the same
	overall = nil
	getMemUsage()
	// Force GC to clear up, should see a memory drop
	runtime.GC()
	getMemUsage()
	fmt.Print("+++++++++++++++++++++++++++++\n")

	// Check CPU Usage
	idle0, total0 := getCPUSample()
	time.Sleep(3 * time.Second)
	idle1, total1 := getCPUSample()
	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
	fmt.Printf("The system CPU usage is: \n %f%% [busy: %f, total: %f]\n",
		cpuUsage, totalTicks-idleTicks, totalTicks)
}

// Check system time
func osTimeCheck() time.Time {
	sysTimeCheck := time.Now()
	return sysTimeCheck
}

// Check system day
func osDayCheck() time.Weekday {
	sysDayCheck := time.Now().Weekday()
	return sysDayCheck
}

// Get uptime since system started
func getUptime() {
	uptime,_ := host.Uptime()
	fmt.Println("Total:", uptime, "seconds")
	days := uptime / (60 * 60 * 24)
	hours := (uptime - (days * 60 * 60 * 24)) / (60 * 60)
	minutes := ((uptime - (days * 60 * 60 * 24))  -  (hours * 60 * 60)) / 60
	fmt.Printf("%d days, %d hours, %d minutes \n",days,hours,minutes)
}

/*
Disk storage evaluation
 */
type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}
const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)
func DiskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}
/*
End of disk storage evaluation
 */

/*
Get memory via evaluation below
 */
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
func getMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}
/*
End of memory evaluation
 */

// Get CPU usage
func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}
