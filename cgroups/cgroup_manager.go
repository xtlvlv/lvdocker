package cgroups

import "modfinal/cgroups/subsystems"

type CgroupManager struct {
	// 限制的值
	Resource *subsystems.ResourceConfig
	// 在当前容器中标识有哪些subsystem需要做限制
	SubsystemsIns []subsystems.Subsystem
}

func (c *CgroupManager) Set()  {
	for _,sub:=range c.SubsystemsIns{
		sub.Set(c.Resource)
	}
}

func (c *CgroupManager) Apply(pid string)  {
	for _,sub:=range c.SubsystemsIns{
		sub.Apply(pid)
	}
}

func (c *CgroupManager) Destroy()  {
	for _,sub:=range c.SubsystemsIns{
		sub.Remove()
	}
}