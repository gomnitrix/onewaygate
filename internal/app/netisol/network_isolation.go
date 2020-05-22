package netisol

import (
	"controller.com/config"
	"controller.com/internal"
	"controller.com/internal/OwmError"
	"github.com/iptables"
)

func NetWorkIsolation(managerID, targetID string) {
	defer OwmError.Pack()
	newIptable, err := iptables.New()
	OwmError.Check(err, "")
	mgrAddr := internal.GetNetInfo(managerID)
	tgtAddr := internal.GetNetInfo(targetID)
	err = newIptable.Append(config.Table, config.PreChain, internal.BuildPreRule(tgtAddr, mgrAddr)...)
	OwmError.Check(err, "")
	err = newIptable.Append(config.Table, config.PostChain, internal.BuildPostRule(tgtAddr, mgrAddr)...)
	OwmError.Check(err, "")
}

func RmTeeRules(managerID, targetID string) {
	defer OwmError.Pack()
	newIptable, err := iptables.New()
	OwmError.Check(err, "")
	mgrAddr := internal.GetNetInfo(managerID)
	tgtAddr := internal.GetNetInfo(targetID)
	err = newIptable.Delete(config.Table, config.PreChain, internal.BuildPreRule(tgtAddr, mgrAddr)...)
	OwmError.Check(err, "")
	err = newIptable.Delete(config.Table, config.PostChain, internal.BuildPostRule(tgtAddr, mgrAddr)...)
	OwmError.Check(err, "")
}
