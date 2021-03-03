package sysinfo

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/godbus/dbus/v5"
	"github.com/vishvananda/netlink"
)

const factor uint64 = 1024 * 1024 // MB
const objName = "org.freedesktop.hostname1"
const objPath = "/org/freedesktop/hostname1"
const getAll = "org.freedesktop.DBus.Properties.GetAll"

type osProperties map[string]dbus.Variant

func (props osProperties) value(key string) string {
	value := props[key].Value()
	if value != nil {
		return value.(string)
	}
	return ""
}

func getHostname(result *SysInfo) {
	hostname, _ := os.Hostname()
	result.Hostname = hostname
}

func getIPAddresses(lnk netlink.Link, result *SysInfo) {
	addrs, err := netlink.AddrList(lnk, 0)
	if err != nil {
		log.Error("Could not read addresses")
		return
	}

	for _, addr := range addrs {
		if addr.IP.DefaultMask() != nil {
			result.IPv4 = addr.IPNet.String()
		} else {
			if !addr.IP.IsLinkLocalUnicast() {
				result.IPv6 = addr.IPNet.String()
			}
		}
	}
}

func getSpaceInfo(baseDir string, result *SysInfo) {
	fd, err := syscall.Open(baseDir, syscall.O_RDONLY, 0x664)
	if err != nil {
		log.Error("%s does not exist", baseDir)
		return
	}
	defer syscall.Close(fd)

	var statfs syscall.Statfs_t
	err = syscall.Fstatfs(fd, &statfs)
	if err != nil {
		log.Error("Failed to get %s file system stats [%s].", baseDir, err)
		return
	}

	// Total Diskspace
	result.SpaceTotal = (uint64(statfs.Bsize) * statfs.Blocks) / factor
	result.SpaceFree = (uint64(statfs.Bsize) * statfs.Bfree) / factor

}

func getMemInfo(result *SysInfo) {
	var info syscall.Sysinfo_t
	if syscall.Sysinfo(&info) == nil {
		result.MemTotal = uint64(info.Totalram) / factor
		result.MemFree = uint64(info.Freeram) / factor
	}
}

func getOSInfo(result *SysInfo) {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Error("Failed to connect to SystemBus bus")
	}
	// Don't close the connection. See:
	// https://github.com/godbus/dbus/issues/144
	osProperties := make(osProperties)
	err = conn.Object(objName, objPath).Call(getAll, 0, objName).Store(&osProperties)
	if err != nil {
		log.WithError(err).Error("Failed to connect to SystemBus bus")
		return
	}

	data, err := ioutil.ReadFile("/sys/firmware/devicetree/base/model")

	if err == nil {
		result.DeviceType = strings.TrimSuffix(string(data), "\u0000")
	}

	data, err = ioutil.ReadFile("/etc/debian_version")

	if err == nil {
		result.Version = strings.TrimSuffix(string(data), "\n")
	}
	kernelName := osProperties.value("KernelName")
	kernelRelease := osProperties.value("KernelRelease")
	result.OS = osProperties.value("OperatingSystemPrettyName")
	result.Kernel = fmt.Sprint(kernelName, " ", kernelRelease)
}

func getMainLink() (netlink.Link, error) {
	list, err := netlink.LinkList()

	if err != nil {
		log.WithError(err).Error("Getting LinkList failed!")
		return nil, err
	}
	for _, lnk := range list {
		name := lnk.Attrs().Name
		if strings.HasPrefix(name, "en") || strings.HasPrefix(name, "eth") || strings.HasPrefix(name, "wl") {
			return lnk, nil
		}
	}
	return nil, errors.New("No matching links found")
}

func getSysInfo() *SysInfo {
	const baseDir = "/var/lib/containerd"
	// const baseDir = "/"
	result := SysInfo{}
	getHostname(&result)
	getSpaceInfo(baseDir, &result)
	getMemInfo(&result)
	getOSInfo(&result)
	getCPUInfo(&result)
	if lnk, err := getMainLink(); err == nil {
		getSerial(lnk, &result)
		getIPAddresses(lnk, &result)
	}
	return &result
}

func getSerial(lnk netlink.Link, result *SysInfo) {
	serial := hex.EncodeToString(lnk.Attrs().HardwareAddr)
	res, err := strconv.ParseUint(serial, 16, 64)
	if err != nil {
		log.WithError(err).Error("Converting MAC failed!")
		return
	}
	result.Serial = strconv.FormatUint(res, 10)
}

func getCPUInfo(result *SysInfo) {
	val, err := ReadValue("/proc/cpuinfo", "model name")
	if err != nil {
		log.Errorf("Error reading value")
		return
	}
	result.Processor = val
}

// ReadValue Reads a file with key/value pairs and converts it to a map
func ReadValue(filename string, value string) (string, error) {
	file, err := os.Open(filename)

	if err != nil {
		log.Printf("failed opening file: %s", err)
		return "", err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, value) {
			parsed := strings.Split(line, ":")
			if len(parsed) == 2 {
				return strings.Trim(parsed[1], " "), nil
			}
		}
	}
	return "", nil
}

func setState(state StateAction) bool {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Error("Failed to connect to SystemBus bus")

	}
	obj := conn.Object("org.freedesktop.login1", "/org/freedesktop/login1").Call("org.freedesktop.login1.Manager."+string(state), 0, false)
	if obj.Err != nil {
		log.WithError(obj.Err).Error("Failed to reboot")
		return false
	}
	return true
}

// StateAction Action to get state
type StateAction string

const (
	// Reboot State Reboot
	Reboot string = "Reboot"
	// PowerOff State PowerOff
	PowerOff string = "PowerOff"
	// Suspend State Suspend
	Suspend string = "Suspend"
	// SuspendThenHibernate State SuspendThenHibernate
	SuspendThenHibernate string = "SuspendThenHibernate"
	// HybridSleep State HybridSleep
	HybridSleep string = "HybridSleep"
	// Hibernate State Hibernate
	Hibernate string = "Hibernate"
	// Halt State Halt
	Halt string = "Halt"
)
