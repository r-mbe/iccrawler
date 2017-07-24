package events

import (
  "github.com/jlaffaye/ftp"
	"github.com/larspensjo/config"
  "strings"
  "os"
  "io"
  "errors"
  "time"
)

type Ftpcli struct {
  Se SyncEntry
  Host string
  Port string
  Username string
  Passwd string
  Path string
  LocalPath string
  File string
  LocalFile string
  Passive bool
  SyncTime string
  st int64
}

func NewFtpcli(se SyncEntry) *Ftpcli {
  conf := se.Conf
  supname := se.Supname
  cli := &Ftpcli{Se: se}

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
    cli.Port = "21"
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
    Warning(se, "conf err: missing file")
  }

  //local-file
  if cnf.HasSection(supname) && cnf.HasOption(supname, "local-file") {
    cli.LocalFile, _ = cnf.String(supname, "local-file")
    cli.LocalFile = FormatFileName(cli.LocalFile)
  } else {
    Warning(se, "conf err: missing local-file")
  }

  //passive
  cli.Passive = false
  if cnf.HasSection(supname) && cnf.HasOption(supname, "passive") {
    passive, _ := cnf.Int(supname, "passive")
    if passive != 0 {
      cli.Passive = true
    }
  }

  //sync-time
  if cnf.HasSection(supname) && cnf.HasOption(supname, "sync-time") {
    cli.SyncTime, _ = cnf.String(supname, "sync-time")
  } else {
    cli.SyncTime = ""
  }

  return cli
}

func (cli *Ftpcli) Pull() (rterr error) {
  rterr = nil
  defer func(){
    if err := recover(); err!=nil {
      Debug(cli.Se, err)
      rterr = errors.New("Recover err")
    }
  }()

  supname := cli.Se.Supname
  Debug(cli.Se, "Check dir:", cli.LocalPath, supname)
	err := os.MkdirAll(cli.LocalPath, 0777)
	if err != nil {
		Warning(cli.Se, "MkdirAll err: ", err, supname)
	}

	Debug(cli.Se, "DialTimeout:", cli.Host, supname)
	conn, err := ftp.DialTimeout(cli.Host+":"+cli.Port, 0, cli.Passive)
	if err != nil {
		Warning(cli.Se, "DialTimeout err:", err)
	}
	defer conn.Quit()

	Debug(cli.Se, "Login:", cli.Username, cli.Passwd, supname)
	err = conn.Login(cli.Username, cli.Passwd)
	if err != nil {
		Warning(cli.Se, "Login err:", err, supname)
	}
	defer conn.Logout()

  Debug(cli.Se, "ChangeDir:", cli.Path, supname)
  err = conn.ChangeDir(cli.Path)
  if err != nil {
    Warning(cli.Se, "ChangeDir err:", err, supname)
  }

	Debug(cli.Se, "List: .", supname)
	entries, err := conn.List(".")
	if err != nil {
		Warning(cli.Se, "List err:", err, supname)
	}

  var size uint64 = 0
  check := false
  newPullTime := ""
  yesterdaySyncTime := cli.Se.GetYesterdaySyncTime()
  lastPullTime := cli.Se.GetLastPullTime()
	for _, entrie := range entries {
    mtime := entrie.Time.Format("2006-01-02 15:04:05")
    mtimeTmp, _ := time.ParseInLocation(timeLayout, mtime, time.Local)
  	if err != nil {
      mtimeTmp = time.Now()
  	}
    nowUnix := time.Now().Unix()
    checkUnix := mtimeTmp.Add(10*time.Minute).Unix()
    Debug(cli.Se, "Pull list:", entrie.Name, entrie.Size, mtime, nowUnix, checkUnix)
    //文件名相同 && 上次更新时间小于文件时间 && 昨天之后更新的数据包
    if entrie.Name == cli.File && yesterdaySyncTime < checkUnix && nowUnix >= checkUnix && lastPullTime < mtimeTmp.Unix() && entrie.Size > 0{
      Normal(cli.Se, "Pull find:", entrie.Name, entrie.Size, mtime)
      size = entrie.Size
      check = true
      newPullTime = mtime
    }
  }

  Debug(cli.Se, "Size:", size, supname)
  if check == false {
    Normal(cli.Se, "Pull Not found:", cli.File, supname)
    return
  }

  Debug(cli.Se, "OpenFile:", cli.LocalPath, cli.LocalFile, supname)
	f, _ := os.OpenFile(cli.LocalPath + cli.LocalFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer f.Close()

  stat, err := f.Stat()
  if err != nil {
		Warning(cli.Se, "Stat err:", err, supname)
	}
  offset := uint64(stat.Size())
  offset = 0
  Debug(cli.Se, "offset:", offset, supname)
  if size > offset {
    Debug(cli.Se, "RetrFrom:", cli.File, offset, supname)
  	r, err := conn.RetrFrom(cli.File, offset)
  	if err != nil {
  		Warning(cli.Se, "RetrFrom err:", err, supname)
  	}
  	defer r.Close()
    len := cli.copy(f, r)

    Normal(cli.Se, "Pull done:", cli.File, len, supname)
  } else {
    Normal(cli.Se, "Pull done yet:", cli.File, 0, supname)
  }

  //UpdateLastSyncTime
  cli.Se.UpdateLastPullTime(newPullTime)
  cli.Se.UpdateLastSyncTime()
  return
}

func (cli *Ftpcli) copy(f *os.File, r io.ReadCloser) int64 {
  supname := cli.Se.Supname
  Debug(cli.Se, "Get Lock: copy", supname)
  lock := getLock("copy:"+supname)
	lock.Lock()
	defer func() {
		lock.Unlock()
		Debug(cli.Se, "Unlocked: copy", supname)
		}()
	Debug(cli.Se, "Locked: copy", supname)

  Debug(cli.Se, "copy: ", cli.File, supname)
  size, err := io.Copy(f, r)
	if err != nil {
		Warning(cli.Se, "Copy err:", err, supname)
	}

  return size
}

func (cli *Ftpcli) Push() (rterr error) {
  rterr = nil
  defer func(){
    if err := recover(); err!=nil {
      Debug(cli.Se, err)
      rterr = errors.New("Recover err")
    }
  }()

  supname := cli.Se.Supname
	Debug(cli.Se, "OpenFile:", cli.LocalPath, cli.LocalFile, supname)
	f, err := os.Open(cli.LocalPath + cli.LocalFile)
  if err != nil {
		Warning(cli.Se, "Push err: ", err, supname)
	}
	defer f.Close()

	Debug(cli.Se, "DialTimeout:", cli.Host, supname)
	conn, err := ftp.DialTimeout(cli.Host+":"+cli.Port, 0, cli.Passive)
	if err != nil {
		Warning(cli.Se, "DialTimeout err:", err)
	}
	defer conn.Quit()

	Debug(cli.Se, "Login:", cli.Username, cli.Passwd, supname)
	err = conn.Login(cli.Username, cli.Passwd)
	if err != nil {
		Warning(cli.Se, "Login err:", err, supname)
	}
	defer conn.Logout()

  Debug(cli.Se, "ChangeDir:", cli.Path, supname)
  err = conn.ChangeDir(cli.Path)
  if err != nil {
    Warning(cli.Se, "ChangeDir err:", err, supname)
  }

	Debug(cli.Se, "List: .", supname)
	entries, err := conn.List(".")
	if err != nil {
		Warning(cli.Se, "List err:", err, supname)
	}

  var offset uint64 = 0
	for _, entrie := range entries {
    if entrie.Name == cli.File {
      Normal(cli.Se, "Push find:", entrie.Name, entrie.Size, entrie.Time.Format("2006-01-02 15:04:05"))
      offset = entrie.Size
    }
  }
  Debug(cli.Se, "offset:", offset, supname)

  stat, err := f.Stat()
  if err != nil {
		Warning(cli.Se, "Stat err:", err, supname)
	}
  size := uint64(stat.Size())
  Debug(cli.Se, "Size:", size, supname)
  if size > offset {
    Debug(cli.Se, "Stor:", cli.File, offset, supname)
  	err := conn.StorFrom(cli.File, f, 0)
  	if err != nil {
  		Warning(cli.Se, "Stor err:", err, supname)
  	}
    Normal(cli.Se, "Push done:", cli.File, size, supname)
  } else {
    Normal(cli.Se, "Push done yet:", cli.File, size, supname)
  }

  //UpdateLastSyncTime
  cli.Se.UpdateLastSyncTime()
  return
}
