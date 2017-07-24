package events

import (
	"github.com/larspensjo/config"
  "strings"
  "errors"
)

type Rsynccli struct {
  Se SyncEntry
  Code string
  Host string
  Port string
  Username string
  Passwd string
  Path string
  LocalPath string
  File string
  LocalFile string
  SyncTime string
}

func NewRsynccli(se SyncEntry) *Rsynccli {
  conf := se.Conf
  supname := se.Supname
  cli := &Rsynccli{Se: se}

  cnf, err := config.ReadDefault(conf)
  if err != nil {
    Warning(se, "conf err:", err)
  }

  //host
  if cnf.HasSection(supname) && cnf.HasOption(supname, "host") {
    cli.Host, _ = cnf.String(supname, "host")
  } else if cnf.HasSection("main") && cnf.HasOption("main", "host") {
    cli.Host, _ = cnf.String("main", "host")
  } else {
    Warning(se, "conf err: missing host")
  }

  //port
  if cnf.HasSection(supname) && cnf.HasOption(supname, "port") {
    cli.Port, _ = cnf.String(supname, "port")
  } else if cnf.HasSection("main") && cnf.HasOption("main", "port") {
    cli.Port, _ = cnf.String("main", "port")
  } else {
    Warning(se, "conf err: missing port")
  }

  //username
  if cnf.HasSection(supname) && cnf.HasOption(supname, "username") {
    cli.Username, _ = cnf.String(supname, "username")
  } else {
    Warning(se, "conf err: missing username")
  }

  //passwd
  if cnf.HasSection(supname) && cnf.HasOption(supname, "passwd") {
    cli.Passwd, _ = cnf.String(supname, "passwd")
  } else {
    Warning(se, "conf err: missing passwd")
  }

  //path
  if cnf.HasSection(supname) && cnf.HasOption(supname, "path") {
    cli.Path, _ = cnf.String(supname, "path")
  } else if cnf.HasSection("main") && cnf.HasOption("main", "path") {
    cli.Path, _ = cnf.String("main", "path")
  } else {
    cli.Path = "./"
  }

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

  cli.Code = `rsync -auvzP \
  --delete \
  --bwlimit=30 \
  --port=` + cli.Port + ` \
  --password-file=`+cli.Passwd+` \
  ` + cli.LocalPath + cli.LocalFile + ` \
  ` + cli.Username + `@` + cli.Host + `:` + cli.Path

  return cli
}

func (cli *Rsynccli) Rsync() (rterr error) {
  Debug(cli.Se, "Rsync:", cli.Se.Supname)
  rterr = nil
  defer func(){
    if err := recover(); err!=nil {
      rterr = errors.New("Recover err")
    }
  }()

  if _, err := Cmd("Rsync", cli.Se, cli.Code, false); err != nil {
    Warning(cli.Se, "Rsync err:", err, " CMD:" + cli.Code)
  }

  //UpdateLastSyncTime
  cli.Se.UpdateLastSyncTime()

  return
}
