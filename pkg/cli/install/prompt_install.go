package install

import (
	"fmt"
	"strings"

	"github.com/jaypipes/ghw"
	"github.com/pterm/pterm"

	"github.com/llmos-ai/llmos/pkg/config"
	"github.com/llmos-ai/llmos/pkg/questions"
	"github.com/llmos-ai/llmos/pkg/utils"
)

const userPlaceHolder = "github:user1,github:user2"

func (i *Installer) AskInstall() error {
	if i.LLMOSConfig.Install.Silent {
		i.logger.Error("Should running in silent mode")
		return nil
	}

	pterm.Info.Println("Welcome to the LLMOS installer")
	install, err := i.LLMOSConfig.Install.DeepCopy()
	if err != nil {
		return fmt.Errorf("failed to create a copy of the install config: %s", err.Error())
	}

	rootDisk, err := AskInstallDevice(install)
	if err != nil {
		if strings.Contains(err.Error(), invalidDeviceNameError) {
			pterm.Error.Println(err.Error())
			return i.AskInstall()
		}
		return err
	}

	if err = AskConfigURL(install); err != nil {
		return err
	}

	osCpy, err := i.LLMOSConfig.OS.DeepCopy()
	if err != nil {
		return fmt.Errorf("failed to create a copy of the OS config: %s", err.Error())
	}

	osCfg, err := AskUserConfigs(osCpy)
	if err != nil {
		return err
	}

	allGood, err := questions.Prompt("Are settings ok?", "n", yesOrNo, true, false)
	if err != nil {
		return err
	}

	if !isYes(allGood) {
		return i.AskInstall()
	}

	// update the config with the user input
	i.LLMOSConfig.Install = *install
	i.LLMOSConfig.OS = *osCfg

	return i.GenerateInstallConfigs(rootDisk)
}

// AskInstallDevice asks the user to choose the installation disk
func AskInstallDevice(install *config.Install) (*ghw.Disk, error) {
	var defaultDevice = "auto"
	var defaultDisk = &ghw.Disk{}
	maxSize := float64(0)

	disks := make(map[string]string)
	block, err := ghw.Block()
	if err == nil {
		for _, disk := range block.Disks {
			// skip useless devices (/dev/ram, /dev/loop, /dev/sr, /dev/zram)
			if strings.HasPrefix(disk.Name, "loop") || strings.HasPrefix(disk.Name, "ram") || strings.HasPrefix(disk.Name, "sr") || strings.HasPrefix(disk.Name, "zram") {
				continue
			}
			diskName := fmt.Sprintf("/dev/%s", disk.Name)
			size := float64(disk.SizeBytes) / float64(GiB)
			if size > maxSize {
				maxSize = size
				defaultDevice = diskName
				defaultDisk = disk
			}
			diskInfo := fmt.Sprintf("%s: %s(%.2f GiB) ", diskName, disk.Model, float64(disk.SizeBytes)/float64(GiB))
			disks[diskName] = diskInfo
		}
	}

	pterm.Info.Println("Available Disks:")
	for _, d := range disks {
		pterm.Info.Println(d)
	}

	device, err := questions.Prompt("Choose the installation disk:", defaultDevice, "Cannot be empty", false, false)
	if err != nil {
		return nil, err
	}

	if disks[device] == "" {
		return nil, fmt.Errorf("%s: %s", invalidDeviceNameError, device)
	}

	install.Device = device

	if err = AskDataDevice(install, disks, device); err != nil {
		return nil, err
	}

	return defaultDisk, nil
}

// AskDataDevice asks the user to choose the data disk
func AskDataDevice(install *config.Install, devices map[string]string, rootDevice string) error {
	prompt := fmt.Sprintf("Use the installation disk(%s)", rootDevice)
	dataDevice, err := questions.Prompt("Choose the data disk:", rootDevice, prompt, true, false)
	if err != nil {
		return err
	}

	if devices[dataDevice] == "" {
		return fmt.Errorf("%s: %s", invalidDeviceNameError, dataDevice)
	}

	if install.Device != dataDevice {
		install.DataDevice = dataDevice
	}

	return nil
}

// AskConfigURL asks the user to provide the LLMOS config file location
func AskConfigURL(install *config.Install) error {
	if install.ConfigURL != "" {
		return nil
	}

	url, err := questions.Prompt("LLMOS config file location (file path or http URL): ", install.ConfigURL, "", true, false)
	if err != nil {
		return err
	}

	install.ConfigURL = url

	if install.ConfigURL != "" {
		if err = utils.IsValidPathOrURL(install.ConfigURL); err != nil {
			return err
		}
	}
	return nil
}

// AskUserConfigs asks the user to provide the user accounting configurations
func AskUserConfigs(os *config.LLMOS) (*config.LLMOS, error) {
	if len(os.SSHAuthorizedKeys) > 0 || os.Password != "" {
		return nil, nil
	}

	username, err := questions.Prompt("User to setup:", defaultLoginUser, emptyPlaceHolder, false, false)
	if err != nil {
		return nil, err
	}

	passwd, err := questions.Prompt("Password:", "", emptyPlaceHolder, false, true)
	if err != nil {
		return nil, err
	}

	users, err := questions.Prompt("SSH authorized keys(optional):", "", userPlaceHolder, true, false)
	if err != nil {
		return nil, err
	}

	// Cleanup the users if we selected the default values as they are not valid users
	if users == userPlaceHolder {
		users = ""
	}

	os.Username = username
	os.Password = passwd
	os.SSHAuthorizedKeys = strings.Split(users, ",")

	return os, nil
}

func isYes(s string) bool {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "y") {
		return true
	}
	return false
}
