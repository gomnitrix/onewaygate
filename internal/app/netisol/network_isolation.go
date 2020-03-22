package netisol

import (
	"controller.com/config"
	"controller.com/internal"
	"github.com/iptables"
)

var log = config.ELog

func NetWorkIsolation(managerID, targetID string) error {
	newIptable, err := iptables.New()
	if err != nil {
		log.Println(err)
		return err
	}
	mgrAddr, err := internal.GetNetInfo(managerID)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	tgtAddr, err := internal.GetNetInfo(targetID)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	err = newIptable.Append(config.Table, config.PreChain, internal.BuildPreRule(tgtAddr, mgrAddr)...)
	if err != nil {
		log.Println(err)
		return err
	}
	err = newIptable.Append(config.Table, config.PostChain, internal.BUildPostRule(tgtAddr, mgrAddr)...)
	if err != nil {
		log.Println(err)
		return err
	}
	//return closeForward(managerID)
	return nil
}

func closeForward(mgrID string) error {
	err := internal.RunCommandInManager(mgrID, []string{"echo", "net.ipv4.ip_forward=0", ">>", "/etc/sysctl.conf"})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func RmTeeRules(managerID, targetID string) error {
	//TODO 重构
	newIptable, err := iptables.New()
	if err != nil {
		log.Println(err)
		return err
	}
	mgrAddr, err := internal.GetNetInfo(managerID)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	tgtAddr, err := internal.GetNetInfo(targetID)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	err = newIptable.Delete(config.Table, config.PreChain, internal.BuildPreRule(tgtAddr, mgrAddr)...)
	if err != nil {
		log.Println(err)
		return err
	}
	err = newIptable.Delete(config.Table, config.PostChain, internal.BUildPostRule(tgtAddr, mgrAddr)...)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
