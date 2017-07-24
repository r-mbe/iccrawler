package events

import (
  "fmt"
  "time"
  "github.com/larspensjo/config"
  "io"
  "os"
  "strings"
  "net"
  "encoding/json"
)

const (
	NONE int = iota
	WARNING
	NORMAL
	DEBUG
)

type Logger struct {
  Level int
  Path string
  NeedRsync bool
  RsyncCD int
  RsyncHost string
  RsyncPort string
  RsyncUsername string
  RsyncPasswd string
  RsyncPath string
  RsyncCode string
  RsyncName string
}

type RsyncEntry struct {
  Timestamp string `json:"@timestamp"`
  Supname string `json:"supname"`
  Action string `json:"action"`
  Message string `json:"message"`
  Ip string `json:"ip"`
  Level int `json:"level"`
  Cost float64 `json:"cost"`
}

var DefLogger *Logger

func NewLogger(cnf *config.Config) *Logger {

  logger := &Logger{
    Level:DEBUG,
    Path:"./",
    NeedRsync: false,
    RsyncCD:5,
    RsyncCode:"",
  }

  if cnf.HasSection("log") && cnf.HasOption("log", "level") {
  	logger.Level, _ = cnf.Int("log", "level")
	}

  if cnf.HasSection("log") && cnf.HasOption("log", "path") {
  	logger.Path, _ = cnf.String("log", "path")
	}
  logger.Path = strings.TrimRight(logger.Path, "/") + "/"

  //RsyncCD
  if cnf.HasSection("log") && cnf.HasOption("log", "rsync-cd") {
  	logger.RsyncCD, _ = cnf.Int("log", "rsync-cd")
    logger.NeedRsync = true
	}

  //RsyncHost
  if cnf.HasSection("log") && cnf.HasOption("log", "rsync-host") {
    logger.RsyncHost, _ = cnf.String("log", "rsync-host")
  } else {
    logger.RsyncHost = ""
  }

  //RsyncPort
  if cnf.HasSection("log") && cnf.HasOption("log", "rsync-port") {
    logger.RsyncPort, _ = cnf.String("log", "rsync-port")
  } else {
    logger.RsyncPort = ""
  }

  //RsyncUsername
  if cnf.HasSection("log") && cnf.HasOption("log", "rsync-username") {
    logger.RsyncUsername, _ = cnf.String("log", "rsync-username")
  } else {
    logger.RsyncUsername = ""
  }

  //RsyncPasswd
  if cnf.HasSection("log") && cnf.HasOption("log", "rsync-passwd") {
    logger.RsyncPasswd, _ = cnf.String("log", "rsync-passwd")
  } else {
    logger.RsyncPasswd = ""
  }

  //RsyncPath
  if cnf.HasSection("log") && cnf.HasOption("log", "rsync-path") {
    logger.RsyncPath, _ = cnf.String("log", "rsync-path")
  } else if logger.NeedRsync {
    panic("log conferr: missing rsync-path")
  }

  //RsyncName
  if cnf.HasSection("log") && cnf.HasOption("log", "rsync-name") {
    logger.RsyncName, _ = cnf.String("log", "rsync-name")
  } else {
    logger.RsyncName = "rsync.log"
  }


  if logger.NeedRsync {
    if logger.RsyncHost != "" {
      logger.RsyncCode = `rsync -auvzP \
      --delete \
      --bwlimit=512 \
      --port=` + logger.RsyncPort + ` \
      --password-file=`+logger.RsyncPasswd+` \
      ` + logger.Path + logger.RsyncName + ` \
      ` + logger.RsyncUsername + `@` + logger.RsyncHost + `:` + logger.RsyncPath
    } else {
      logger.RsyncCode = "cp " + logger.Path + logger.RsyncName + " " + logger.RsyncPath  + logger.RsyncName
    }
  }

  return logger
}

func (l Logger) log(level int, msg ...interface{}) {
	if l.Level >= level {
		now := time.Now().Format("2006-01-02 15:04:05 ")
		fmt.Print(now)
		fmt.Println(msg...)
	}
}

func (l Logger) Normal(msg ...interface{}) {
	l.log(NORMAL, msg...)
}

func (l Logger) Warning(msg ...interface{}) {
	l.log(WARNING, msg...)
  now := time.Now().Format("2006-01-02 15:04:05 ")
  err := fmt.Sprintf("%s %v\n", now, msg)
  l.WriteFile("error.log", []byte(err))
  panic(msg)
}

func (l Logger) Debug(msg ...interface{}) {
	l.log(DEBUG, msg...)
}

func (l Logger) ReadyRsync(msg string, lvl int, seval interface{}) {
  if len(l.RsyncName) > 0 && seval != nil {
    //构建同步日志
    se, _ := seval.(SyncEntry)
    parts := strings.Split(se.Supname, "_")
    supname := parts[0]
    now := time.Now().Format(time.RFC3339)
    ip := getIp()
    cost := se.GetCostTime()
    data := &RsyncEntry{
      Timestamp: now,
      Supname: supname,
      Action: se.Method,
      Message: msg,
      Ip: ip,
      Level: lvl,
      Cost: cost,
    }

    //Marshal
    jsondata, err := json.Marshal(data)
    if err != nil {
      l.Warning("ReadyRsync json.Marshal err:", err)
    }
    jsondata = append(jsondata, '\n')
    //WriteFile
    l.WriteFile(l.RsyncName, jsondata)
  }
}

func getIp() string {
  ips := []string{}
  addrs, err := net.InterfaceAddrs()
	if err == nil {
    for _, addr := range addrs {
  		if ipnet, ok := addr.(*net.IPNet); ok {
        if ipnet.IP.IsLoopback() {
          continue
        }
  			if ipnet.IP.To4() != nil {
          ips = append(ips, ipnet.IP.String())
  			}
  		}
  	}
	}

  return strings.Join(ips, " ")
}

func (l Logger) Rsync() {
  if !l.NeedRsync {
    l.Debug("Log rsync: close")
    return
  }
  l.Debug("Log rsync: start..")
  for {
    if _, err := Cmd("Log", nil, l.RsyncCode, false); err != nil {
      l.Debug("Log rsync err:", err, " CMD:" + l.RsyncCode)
    }
    time.Sleep(time.Duration(l.RsyncCD) * time.Second)
  }
}

func (l Logger) WriteFile(filename string, data []byte) error {
  f, err := os.OpenFile(l.Path + filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
  if err != nil {
    return err
  }
  defer f.Close()

  n, err := f.Write(data)
  if err == nil && n < len(data) {
    err = io.ErrShortWrite
  }
  return err
}
