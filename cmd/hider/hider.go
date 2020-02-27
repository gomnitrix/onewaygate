package hider

import (
	"os"
	"strings"

	"controller.com/config"

	"controller.com/internal"
	"github.com/codeskyblue/go-sh"
)

func Hide(hostPids []string) {
	fp, err := os.OpenFile("/home/fp.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return
	}
	defer fp.Close()

	out, err := sh.Command("ps", "-aux").Command("awk", "{print $2}").Output()
	if err != nil {
		panic(err)
	}
	outFmt := string(out)
	pidsInCont := strings.Split(outFmt, "\n")
	_, err = fp.WriteString(outFmt + "\n")
	if err != nil {
		panic(err)
	}
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
			_, err = fp.WriteString("pid in host: " + truePid + "---" + "pid in cont: " + pid)
			if err != nil {
				panic("line42:" + err.Error())
			}
			hideOut, err := sh.Command(config.AvaPath, "i", pid).Output()
			if err != nil {
				fp.WriteString("error:" + err.Error())
			} else {
				fp.WriteString("bash output:" + string(hideOut))
			}
		}
	}
}
