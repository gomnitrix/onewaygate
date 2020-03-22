package netisol

import (
	"controller.com/config"
	"controller.com/internal"

	"github.com/iptables"
)

var log = config.ELog
var IPTables = iptables.IPTables{}

func NetWorkIsolation(managerID, targetID string) error {
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
	err = IPTables.Append(config.Table, config.PreChain, internal.BuildPreRule(tgtAddr, mgrAddr))
	if err != nil {
		log.Println(err)
		return err
	}
	err = IPTables.Append(config.Table, config.PostChain, internal.BUildPostRule(tgtAddr, mgrAddr))
	if err != nil {
		log.Println(err)
		return err
	}
	return closeForward(mgrAddr)
}

func closeForward(mgrID string) error {
	err := internal.RunCommandInManager(mgrID, []string{"sysctl", "net.ipv4.ip_forward", "=", "0"})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
