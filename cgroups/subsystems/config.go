package subsystems

const (
	CgroupDirName = "lvdocker"
)

type ResourceConfig struct {
	MemoryLimit string
	CpuLimit string
}

type Subsystem interface {
	Name() string
	Set(res *ResourceConfig) error
	Apply(pid string) error
	Remove() error
}