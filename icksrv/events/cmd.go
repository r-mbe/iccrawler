package events

import (
  "github.com/larspensjo/config"
  "strings"
  "os"
  "os/exec"
  "time"
  "errors"
  "fmt"
)

type Cmdcli struct {
  Se SyncEntry
  Condition string
  Code string
  Path string
  LocalPath string
  File string
  LocalFile string
  SyncTime string
}

func NewCmdcli(se SyncEntry) *Cmdcli {
  conf := se.Conf
  supname := se.Supname
  cli := &Cmdcli{Se: se}

  cnf, err := config.ReadDefault(conf)
  if err != nil {
    Warning(se, "conf err:", err)
  }

  //path
  if cnf.HasSection(supname) && cnf.HasOption(supname, "path") {
    cli.Path, _ = cnf.String(supname, "path")
  } else if cnf.HasSection("main") && cnf.HasOption("main", "path") {
    cli.Path, _ = cnf.String("main", "path")
  } else {
    cli.Path = "./"
  }
  cli.Path = strings.TrimRight(cli.Path, "/") + "/"

  //local-path
  if cnf.HasSection(supname) && cnf.HasOption(supname, "local-path") {
    cli.LocalPath, _ = cnf.String(supname, "local-path")
  } else if cnf.HasSection("main") && cnf.HasOption("main", "local-path") {
    cli.LocalPath, _ = cnf.String("main", "local-path")
  } else {
    cli.LocalPath = "./"
  }
  cli.LocalPath = strings.TrimRight(cli.LocalPath, "/") + "/"

  //file
  if cnf.HasSection(supname) && cnf.HasOption(supname, "file") {
    cli.File, _ = cnf.String(supname, "file")
    cli.File = FormatFileName(cli.File)
  } else {
    cli.File = ""
  }

  //local-file
  if cnf.HasSection(supname) && cnf.HasOption(supname, "local-file") {
    cli.LocalFile, _ = cnf.String(supname, "local-file")
    cli.LocalFile = FormatFileName(cli.LocalFile)
  } else {
    cli.LocalFile = ""
  }

  //sync-time
  if cnf.HasSection(supname) && cnf.HasOption(supname, "sync-time") {
    cli.SyncTime, _ = cnf.String(supname, "sync-time")
  } else {
    cli.SyncTime = ""
  }

  //Condition
  if cnf.HasSection(supname) && cnf.HasOption(supname, "condition") {
    cli.Condition, _ = cnf.String(supname, "condition")
  } else {
    cli.Condition = ""
  }

  //Code
  if cnf.HasSection(supname) && cnf.HasOption(supname, "code") {
    cli.Code, _ = cnf.String(supname, "code")
    cli.Code = FormatFileName(cli.Code)
  } else {
    Warning(se, "conf err: missing code")
  }

  //format
  r := strings.NewReplacer([]string{"$local-path", cli.LocalPath, "$local-file", cli.LocalFile, "$path", cli.Path, "$file", cli.File }...)
	cli.Code = r.Replace(cli.Code)

  return cli
}

func Cmd(action string, se interface{}, command string, retry bool) (string, error) {
  supname := ""
  if sev, ok := se.(SyncEntry); ok {
    supname = sev.Supname
  }

  Normal(se, action, "begin, CMD:", command, supname)

  Debug(se, action, "Cmd Lock: ", supname)
  lock := getLock("cmd:" + supname)
  lock.Lock()
	defer func() {
		lock.Unlock()
		Debug(se, action, "Unlocked: Cmd", supname)
	}()
  Debug(se, action, "Cmd Locked: ", supname)

  times := 1
  maxTimes := 5
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.CombinedOutput()

	for err != nil && retry {
		time.Sleep(1 * time.Second)
    out, err = cmd.CombinedOutput()
    if times >= maxTimes {
      break
    }
    times = times + 1
	}

	if err != nil {
    errmsg := fmt.Sprintf("%s %s",string(out), err.Error())
    err = errors.New(errmsg)
		return "", err
	}

  Normal(se, action, "done, CMD:", command)

	return string(out), nil
}

func (cli *Cmdcli) Exec(method string) (out string, rterr error) {
  out = ""
  rterr = nil
  defer func(){
    if err := recover(); err!=nil {
      Debug(cli.Se, err)
      rterr = errors.New("Recover err")
    }
  }()

  //Condition
  switch cli.Condition {
    case "mtime":
      f, _ := os.Open(cli.LocalPath + cli.LocalFile)
      defer f.Close()
      stat, err := f.Stat()
      if err != nil {
    		Warning(cli.Se, "Stat err:", err, cli.Se.Supname)
    	}

      mtime := stat.ModTime().Format(timeLayout)
      mtimeTmp, _ := time.ParseInLocation(timeLayout, mtime, time.Local)
    	if err != nil {
        mtimeTmp = time.Now()
    	}
      nowUnix := time.Now().Unix()
      checkUnix := mtimeTmp.Add(10*time.Minute).Unix()

      if mtimeTmp.Unix() <= cli.Se.GetYesterdaySyncTime() || checkUnix <= cli.Se.GetLastSyncTime() {
        Debug(cli.Se, method, "mtime:", mtime, cli.LocalFile, cli.Se.Supname, mtimeTmp.Unix(), cli.Se.GetYesterdaySyncTime())
        rterr = errors.New("mtime not find" + cli.LocalFile)
        return
      }

      if checkUnix > nowUnix {
        Debug(cli.Se, method, "mtime wait:", mtimeTmp.Add(10*time.Minute).Format(timeLayout), cli.LocalFile, cli.Se.Supname)
        rterr = errors.New("mtime find " + cli.LocalFile + " wait..")
        return
      }

      Normal(cli.Se, method, "mtime ok:", mtime, cli.LocalFile, cli.Se.Supname)
    default:
  }

  supname := cli.Se.Supname
	err := os.MkdirAll(cli.Path, 0777)
	if err != nil {
		Warning(cli.Se, method, "MkdirAll err: ", err, supname)
	}

  out, err = Cmd("Sh", cli.Se, cli.Code, false)
  if err != nil {
    Warning(cli.Se, method, "err:", err, " CMD:" + cli.Code)
  }

  return
}

func (cli *Cmdcli) Sh() error {
  Debug(cli.Se, "Sh:", cli.Se.Supname)
  out, err := cli.Exec("Sh");

  if err != nil {
    return err
  }

  Normal(cli.Se, "Sh succ:", out, cli.Se.Supname)

  //UpdateLastSyncTime
  cli.Se.UpdateLastSyncTime()

  return nil
}

func (cli *Cmdcli) Import() error {
  Debug(cli.Se, "Import:", cli.Se.Supname)
  out, err := cli.Exec("Import");

  if err != nil {
    return err
  }

  if !strings.Contains(out, "succ") {
    Normal(cli.Se, "Import err:", out, cli.Se.Supname)
    return errors.New("Import err:" + out)
  } else {
    Normal(cli.Se, "Import succ:", out, cli.Se.Supname)
  }

  //UpdateLastSyncTime
  cli.Se.UpdateLastSyncTime()

  return nil
}
