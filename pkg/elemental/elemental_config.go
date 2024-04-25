package elemental

import (
	"fmt"
	"path/filepath"

	"github.com/jaypipes/ghw"
	elconst "github.com/rancher/elemental-toolkit/pkg/constants"

	"github.com/llmos-ai/llmos/pkg/config"
	"github.com/llmos-ai/llmos/pkg/constants"
)

const (
	SoftMinDiskSizeGiB   = 140
	HardMinDiskSizeGiB   = 80
	MinCosPartSizeGiB    = 25
	NormalCosPartSizeGiB = 60
)

type ElementalConfig struct {
	Install  InstallSpec `yaml:"install,omitempty"`
	Reboot   bool        `yaml:"reboot,omitempty"`
	Poweroff bool        `yaml:"poweroff,omitempty"`
}

type InstallSpec struct {
	Target          string            `yaml:"target,omitempty"`
	Partitions      *DefaultPartition `yaml:"partitions,omitempty"`
	ExtraPartitions []Partition       `yaml:"extra-partitions,omitempty"`
	ISO             string            `yaml:"iso,omitempty"`
	CloudInit       string            `yaml:"cloud-init,omitempty"`
	System          string            `yaml:"system,omitempty"`
	TTY             string            `yaml:"tty,omitempty"`
}

type DefaultPartition struct {
	OEM        *Partition `yaml:"oem,omitempty"`
	State      *Partition `yaml:"state,omitempty"`
	Recovery   *Partition `yaml:"recovery,omitempty"`
	Persistent *Partition `yaml:"persistent,omitempty"`
}

type Partition struct {
	FilesystemLabel string `yaml:"label,omitempty"`
	Size            uint   `yaml:"size,omitempty"`
	FS              string `yaml:"fs,omitempty"`
}

func NewElementalConfig(path string, i config.Install) *ElementalConfig {
	cfg := &ElementalConfig{
		Install: InstallSpec{
			Target: path,
		},
		Reboot:   i.Reboot,
		Poweroff: i.PowerOff,
	}

	if i.ConfigURL != "" {
		cfg.Install.CloudInit = i.ConfigURL
	}

	if i.TTY != "" {
		cfg.Install.TTY = i.TTY
	}

	if i.SystemURI != "" {
		cfg.Install.System = i.SystemURI
	}

	if i.ISOUrl != "" {
		cfg.Install.ISO = i.ISOUrl
	}

	return cfg
}

func GenerateElementalConfig(cfg *config.LLMOSConfig, rootDisk *ghw.Disk) (*ElementalConfig, error) {
	path, err := filepath.EvalSymlinks(cfg.Install.Device)
	if err != nil {
		return nil, err
	}
	elementalConfig := NewElementalConfig(path, cfg.Install)

	//customize data partition layout
	elementalConfig, err = CreateRootPartitioningLayout(cfg, elementalConfig, rootDisk)
	if err != nil {
		return nil, err
	}

	return elementalConfig, nil
}

func CreateRootPartitioningLayout(cfg *config.LLMOSConfig, elementalConfig *ElementalConfig, rootDisk *ghw.Disk) (*ElementalConfig, error) {
	var err error
	cosPersistentSizeGiB := uint64(0)
	if cfg.HasDataPartition() {
		diskSizeBytes := rootDisk.SizeBytes
		cosPersistentSizeGiB, err = calcCosPersistentPartSize(diskSizeBytes >> 30)
		if err != nil {
			return nil, err
		}
		cosPersistentSizeGiB = cosPersistentSizeGiB << 10
	}

	elementalConfig.Install.Partitions = &DefaultPartition{
		OEM: &Partition{
			FilesystemLabel: elconst.OEMLabel,
			Size:            elconst.OEMSize,
			FS:              elconst.LinuxFs,
		},
		State: &Partition{
			FilesystemLabel: elconst.StateLabel,
			Size:            constants.StateSize, // adding more size for air-gap images
			FS:              elconst.LinuxFs,
		},
		Recovery: &Partition{
			FilesystemLabel: elconst.RecoveryLabel,
			Size:            constants.RecoverySize, // ditto
			FS:              elconst.LinuxFs,
		},
		Persistent: &Partition{
			FilesystemLabel: elconst.PersistentLabel,
			Size:            uint(cosPersistentSizeGiB),
			FS:              elconst.LinuxFs,
		},
	}

	if cfg.HasDataPartition() {
		elementalConfig.Install.ExtraPartitions = []Partition{
			{
				FilesystemLabel: "LLMOS_DATA_PERSISTENT",
				Size:            0,
				FS:              elconst.LinuxFs,
			},
		}
	}

	return elementalConfig, nil
}

func calcCosPersistentPartSize(diskSizeGiB uint64) (uint64, error) {
	switch {
	case diskSizeGiB < HardMinDiskSizeGiB:
		return 0, fmt.Errorf("disk too small: %dGB. Minimum %dGB is required", diskSizeGiB, HardMinDiskSizeGiB)
	case diskSizeGiB < SoftMinDiskSizeGiB:
		d := MinCosPartSizeGiB / float64(SoftMinDiskSizeGiB-HardMinDiskSizeGiB)
		partSizeGiB := MinCosPartSizeGiB + float64(diskSizeGiB-HardMinDiskSizeGiB)*d
		return uint64(partSizeGiB), nil
	default:
		partSizeGiB := NormalCosPartSizeGiB + ((diskSizeGiB-100)/100)*10
		if partSizeGiB > 100 {
			partSizeGiB = 100
		}
		return partSizeGiB, nil
	}
}
