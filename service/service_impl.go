package service

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/coreos/go-systemd/v22/unit"
)

// ListUnit returns status information of the specified unit
func ListUnit(name string) (dbus.UnitStatus, error) {
	var status dbus.UnitStatus

	con, err := dbus.NewSystemConnection()
	defer con.Close()
	if err != nil {
		return status, err
	}

	units, err := con.ListUnitsByNames([]string{name})
	if err != nil {
		return status, err
	}

	return units[0], nil
}

// DisableUnitFile may be used to disable the specified unit by removing symlinks to it
func DisableUnitFile(name string) error {
	log.Info("Disable unit file " + name)
	con, err := dbus.NewSystemConnection()
	defer con.Close()
	if err != nil {
		return err
	}

	_, err = con.DisableUnitFiles([]string{name}, false)
	return nil
}
func ResetFailedUnitFile(name string) error {
	log.Info("reset failed unit file " + name)
	con, err := dbus.NewSystemConnection()
	defer con.Close()
	if err != nil {
		return err
	}

	err = con.ResetFailedUnit(name)
	return err
}

// EnableUnitFile may be used to enable the specified unit by creating symlinks to it
func EnableUnitFile(name string) error {
	log.Info("Enable unit file " + name)
	con, err := dbus.NewSystemConnection()
	defer con.Close()
	if err != nil {
		return err
	}
	_, _, err = con.EnableUnitFiles([]string{name}, false, false)
	return err
}

// StartUnit enqueues a start job and depending jobs, if any (unless otherwise specified by the mode string)
func StartUnit(name string, mode string) error {
	log.Infof("Start unit %s", name)
	con, err := dbus.NewSystemConnection()
	defer con.Close()
	if err != nil {
		return err
	}
	reschan := make(chan string)
	_, err = con.StartUnit(name, mode, reschan)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	job := <-reschan
	if job != "done" {
		return errors.New("Job is not done:" + job)

	}
	time.Sleep(1 * time.Second)
	unit := getUnitStatus(con, name)
	if unit == nil {
		return errors.New(name + " not found")
	} else if unit.ActiveState != "active" {
		return errors.New(name + " not active")
	}
	return nil
}

// StopUnit stops the specified
func StopUnit(name string, mode string) error {
	log.Infof("Stop unit " + name)
	con, err := dbus.NewSystemConnection()
	defer con.Close()
	if err != nil {
		return err
	}
	ch := make(chan string)
	_, err = con.StopUnit(name, mode, ch)
	if err != nil {
		return err
	}
	recv := <-ch
	log.Info("Result of op: ", recv)

	return nil
}

// RestartUnit restarts an unit
func RestartUnit(name string, mode string) error {
	log.Infof("Restart unit " + name)
	con, err := dbus.NewSystemConnection()
	defer con.Close()
	if err != nil {
		return err
	}
	ch := make(chan string)
	_, err = con.RestartUnit(name, mode, ch)

	if err != nil {
		return err
	}
	job := <-ch
	if job != "done" {
		return errors.New("Job is not done:" + job)

	}
	time.Sleep(1 * time.Second)
	unit := getUnitStatus(con, name)
	if unit == nil {
		return errors.New(name + " doesn't exist")
	} else if unit.ActiveState != "active" {
		return errors.New(name + " not active")
	}

	return nil
}

// ReadUnitFile reads the specified unit file
func ReadUnitFile(filename string) ([]*unit.UnitOption, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return unit.DeserializeOptions(f)
}

// WriteUnitFile writes an unit file
func WriteUnitFile(filename string, templateName string, values ...string) error {
	fileBytes, err := ioutil.ReadFile("templates/" + templateName + ".json")
	if err != nil {
		log.Error("Template not found")
		return errors.New("Template not found")
	}
	fileString := string(fileBytes)
	if len(values) > 0 {
		r := strings.NewReplacer(values...)
		fileString = r.Replace(fileString)
	}
	var options []*unit.UnitOption

	if err := json.Unmarshal([]byte(fileString), &options); err != nil {
		log.WithError(err).Error("Parse error")
		return err
	}

	reader := unit.Serialize(options)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Error(err)
	}
	// create temporary file
	tmpFile, err := ioutil.TempFile("", "*")
	if err != nil {
		log.Error(err)
	}
	defer os.Remove(tmpFile.Name()) // clean up
	if _, err := tmpFile.Write(content); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		log.Error(err)
	}
	// check if filename already exists
	log.Tracef("Checking %s", filename)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Tracef("%s does not exist ", filename)
		err := os.Rename(tmpFile.Name(), filename)
		if err != nil {
			return err
		}
		err = os.Chmod(filename, 0755)
		if err != nil {
			return err
		}
	} else if err == nil {
		log.Tracef("%s does exist ", filename)
		// compare checksum
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		tmpData, err := ioutil.ReadFile(tmpFile.Name())
		if err != nil {
			return err
		}
		dataSum := md5.Sum(data)
		tmpDataSum := md5.Sum(tmpData)
		log.Tracef("Checksum of %s: %x", filename, dataSum)
		log.Tracef("Checksum of %s: %x", tmpFile.Name(), tmpDataSum)
		if dataSum != tmpDataSum {
			log.Tracef("Overwriting %s", filename)
			err := rename(tmpFile.Name(), filename)
			if err != nil {
				return err
			}
		}
	} else {
		log.Tracef("%s may exist ", filename)
		err := rename(tmpFile.Name(), filename)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReloadDaemon() error {
	con, err := dbus.NewSystemConnection()
	defer con.Close()
	err = con.Reload()
	return err
}

func rename(old string, new string) error {
	err := os.Rename(old, new)
	if err != nil {
		return err
	}

	con, err := dbus.NewSystemConnection()
	defer con.Close()
	if err != nil {
		return err
	}
	// systemctl reload
	err = con.Reload()
	if err != nil {
		log.Warn(err)
	}

	return nil
}

// DeleteUnit deletes the specified unit file
func DeleteUnit(name string) error {
	log.Info("Delete unit " + name)
	if err := StopUnit(name, "fail"); err != nil {
		log.Error(err.Error())
	}
	if err := DisableUnitFile(name); err != nil {
		log.Error(err.Error())
	}
	if err := ResetFailedUnitFile(name); err != nil {
		log.Error(err.Error())
	}
	if err := os.Remove("/etc/systemd/system/" + name); err != nil {
		return err
	}
	return nil
}

func getUnitStatus(conn *dbus.Conn, name string) *dbus.UnitStatus {
	units, err := conn.ListUnits()
	if err != nil {
		return nil
	}
	for _, u := range units {
		if u.Name == name {
			return &u
		}
	}
	return nil
}

func copy(src string, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		log.Error("Failed to open %s: %s", src, err)
		return err
	}
	defer input.Close()
	output, err := os.Create(dst)
	if err != nil {
		log.Error("Failed to create %s: %s", dst, err)
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, input)
	if err != nil {
		log.Error("Failed to copy %s to %s: %s", src, dst, err)
		return err
	}
	err = output.Sync()
	if err != nil {
		log.Error("Failed to sync %s: %s", dst, err)
		return err
	}
	return nil
}

func upgradeBoot(mountOptions string) (string, error) {
	dir, err := ioutil.TempDir(os.TempDir(), "*")
	if err != nil {
		return "", err
	}
	defer os.Remove(dir)
	exp := regexp.MustCompile(`(^| )root=[^ \n]*`)
	data, err := ioutil.ReadFile(filepath.Join("/", "proc", "cmdline"))
	if err != nil {
		log.Error("Failed to read cmdline: %s", err)
		return "", err
	}
	var slice = exp.Find(data)
	if slice == nil {
		log.Error("Failed to find root in %s", data)
		return "", err
	}
	var root = strings.TrimPrefix(string(slice), " ")
	root = strings.TrimPrefix(root, "root=")
	if root == "/dev/nfs" {
		data, err = ioutil.ReadFile(filepath.Join("/", "etc", "fstab"))
		if err != nil {
			log.Error("Failed to read fstab: %s", err)
			return "", err
		}
		exp = regexp.MustCompile(`^/dev/[^\s]+`)
		var slice = exp.Find(data)
		if slice == nil {
			log.Error("Failed to find dev in %s", data)
			return "", err
		}
		root = string(slice)
	}
	stat := syscall.Stat_t{}
	syscall.Stat(root, &stat)
	for i := uint64(stat.Rdev%256) - 1; i > 0; i-- {
		dev, err := ioutil.TempFile("/dev", "*")
		if err != nil {
			log.Error("Failed to create dev: %s", err)
			return "", err
		}
		os.Remove(dev.Name())
		err = syscall.Mknod(dev.Name(), syscall.S_IFBLK|0644, int(stat.Rdev-i))
		if err != nil {
			log.Error("Failed to create dev %s: %s", dev.Name(), err)
			return "", err
		}
		err = syscall.Mount(dev.Name(), dir, "vfat", 0, "")
		if err != nil {
			log.Error("Failed to mount %s at %s: %s", dev.Name(), dir, err)
			return "", err
		}
		defer syscall.Unmount(dir, 0)

		data, err = ioutil.ReadFile("/sys/firmware/devicetree/base/model")
		switch {
		case strings.HasPrefix(string(data), "Raspberry Pi "):
			log.Info("Detected: 'Raspberry Pi'")
			_, err = os.Stat(filepath.Join(dir, "cmdline.txt"))
			if os.IsNotExist(err) {
				log.Error("Failed to find cmdline.txt: %s", err)
				return "", err
			}
			tmp, err := ioutil.TempFile(dir, "*")
			if err != nil {
				log.Error("Failed to create temp file: %s", err)
				return "", err
			}
			defer func() {
				tmp.Close()
				os.Remove(tmp.Name())
			}()
			_, err = fmt.Fprintf(tmp, "boot=overlay root=%s rw apparmor=1 security=apparmor\n", root)
			if err != nil {
				log.Error("Failed to write cmdline.txt: %s", err)
				return "", err
			}
			err = os.Rename(tmp.Name(), filepath.Join(dir, "cmdline.txt"))
			if err != nil {
				log.Error("Failed to rename cmdline.txt: %s", err)
				return "", err
			}
		case strings.HasPrefix(string(data), "HARTING "):
			log.Info("Detected: 'HARTING'")
			tmp, err := ioutil.TempFile(dir, "*")
			if err != nil {
				log.Error("Failed to create temp file: %s", err)
				return "", err
			}
			defer func() {
				tmp.Close()
				os.Remove(tmp.Name())
			}()

			err = copy("/boot/vmlinuz", filepath.Join(dir, "vmlinuz.new"))
			if err != nil {
				return "", err
			}

			err = copy("/boot/initrd.uImage", filepath.Join(dir, "initrd.uImage.new"))
			if err != nil {
				return "", err
			}

			err = copy("/boot/dtb", filepath.Join(dir, "dtb.new"))
			if err != nil {
				return "", err
			}

			data, err = ioutil.ReadFile("/boot/uboot.env")
			if err != nil {
				log.Error("Failed to read uboot.env: %s", err)
				return "", err
			}

			err = ioutil.WriteFile(filepath.Join(dir, "uboot.env.new"), exp.ReplaceAll(data, []byte(fmt.Sprintf(" root=%s", root))), 0644)
			if err != nil {
				log.Error("Failed to write uboot.env.new: %s", err)
				return "", err
			}

			err = os.Rename(filepath.Join(dir, "vmlinuz.new"), filepath.Join(dir, "vmlinuz"))
			if err != nil {
				log.Error("Failed to rename vmlinuz: %s", err)
				return "", err
			}
			err = os.Rename(filepath.Join(dir, "initrd.uImage.new"), filepath.Join(dir, "initrd.uImage"))
			if err != nil {
				log.Error("Failed to rename initrd.uImage: %s", err)
				return "", err
			}
			err = os.Rename(filepath.Join(dir, "dtb.new"), filepath.Join(dir, "dtb"))
			if err != nil {
				log.Error("Failed to rename dtb: %s", err)
				return "", err
			}
			err = os.Rename(filepath.Join(dir, "uboot.env.new"), filepath.Join(dir, "uboot.env"))
			if err != nil {
				log.Error("Failed to rename uboot.env: %s", err)
				return "", err
			}
		}
	}
	return root, nil
}

func upgradeRoot(hostname string, root string, mountOptions string) error {
	dir, err := ioutil.TempDir(os.TempDir(), "*")
	if err != nil {
		log.Error("Failed to create dir: %s", err)
		return err
	}
	defer os.Remove(dir)
	cmd := exec.Command("lsblk", "-n", "-o", "FSTYPE", root)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Error("Failed to detect fstype: %s", err)
		return err
	}
	err = syscall.Mount(root, dir, strings.TrimSuffix(out.String(), "\n"), 0, "")
	if err != nil {
		log.Error("Failed to mount %s at %s: %s", root, dir, err)
		return err
	}
	defer syscall.Unmount(dir, 0)
	err = ioutil.WriteFile(filepath.Join(dir, "rootflags.new"), []byte(fmt.Sprintf("-o %s", mountOptions)), 0644)
	if err != nil {
		log.Error("Failed to write rootflags: %s", err)
		return err
	}
	os.Rename(filepath.Join(dir, "rootflags.new"), filepath.Join(dir, "rootflags"))
	if err != nil {
		log.Error("Failed to rename rootflags: %s", err)
		return err
	}
	if len(hostname) > 0 {
		var hostpath = filepath.Join(dir, "runtime", "fs", "etc")
		_, err := os.Stat(hostpath)
		if err != nil {
			os.MkdirAll(hostpath, 0755)
			os.MkdirAll(filepath.Join(dir, "runtime", "work"), 0755)
		}
		err = ioutil.WriteFile(filepath.Join(hostpath, "hostname.new"), []byte(hostname), 0644)
		log.Debug("HOSTNAME ", hostname)
		if err != nil {
			log.Error("Failed to write hostname: %s", err)
			return err
		}
		os.Rename(filepath.Join(hostpath, "hostname.new"), filepath.Join(hostpath, "hostname"))
		if err != nil {
			log.Error("Failed to rename hostname: %s", err)
			return err
		}
	}
	return nil
}
func upgrade(hostname string, mountOptions string) error {
	exp := regexp.MustCompile(`workdir=[^,]+,upperdir=[^,]+`)
	mountOptions = exp.ReplaceAllString(mountOptions, "workdir=/mnt/runtime/work,upperdir=/mnt/runtime/fs")
	root, err := upgradeBoot(mountOptions)
	if err != nil {
		return err
	}
	err = upgradeRoot(hostname, root, mountOptions)

	return err
}
