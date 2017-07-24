package events

import (
  "fmt"
  "time"
  "errors"
  "github.com/larspensjo/config"
  "strings"
  "strconv"
  "regexp"
  "io/ioutil"
  . "sync"
)

type CheckEntry struct {
  Conf string
  Supname string
}

type SyncEntry struct {
  Conf  string
  Supname string
  Code string
  LogPath string
  Method string
  SyncTime string
  SyncEndTime string
  SyncNext map[string]SyncEntry
  LastSyncTime int64
  st int64
}

type EventEntry struct {
  Name string
  Se SyncEntry
}

type Event func(EventEntry)

var deflock = &RWMutex{}
var locks = make(map[string]*RWMutex)
var timeLayout = "2006-01-02 15:04:05"
var chEvent = make(chan EventEntry, 1)

func getLock(key string) (lock *RWMutex) {
  deflock.Lock()
  defer func() {
		deflock.Unlock()
	}()

  ok := false
  if lock, ok = locks[key]; ok {
    return
  }

  locks[key] = &RWMutex{}
  lock, _ = locks[key]

  return
}

func FormatFileName(name string) string {
  timeFormat := make([]string, 8)
  now := time.Now()
  day := []byte(now.Format(timeLayout)) // 2016-10-10 11:11:11
  reg := regexp.MustCompilePOSIX(`%(\-?[0-9]+)d`)
  if matches := reg.FindStringSubmatch(name); matches!=nil {
    if diff, err := strconv.Atoi(matches[1]); err == nil {
      day = []byte(now.AddDate(0,0,diff).Format(timeLayout)) // 2016-10-10 11:11:11
      timeFormat = append(timeFormat, matches[0], string(day[8:10]))
    }
  }
  // fmt.Println(string(now))
  timeFormat = append(timeFormat,
    "%Y", string(day[:4]),
    "%y", string(day[2:4]),
    "%m", string(day[5:7]),
    "%d", string(day[8:10]),
    "%H", string(day[11:13]),
    "%i", string(day[14:16]),
    "%s", string(day[17:19]),
  )
	r := strings.NewReplacer(timeFormat...)
	formatted := r.Replace(name)
	return formatted
}

func Pull(ee EventEntry) {
  Debug(ee.Se, "Pull Lock: ", ee.Name, ee.Se.Supname)
  lock := getLock("Pull:"+ee.Se.Supname)
  lock.Lock()
	defer func() {
		lock.Unlock()
		Debug(ee.Se, "Unlocked: Pull", ee.Name, ee.Se.Supname)
	}()
  Debug(ee.Se, "Pull Locked: ", ee.Name, ee.Se.Supname)

  syncTime, err := ee.Se.GetCurrentSyncTime()
  if err != nil {
    Normal(ee.Se, "Pull se.GetCurrentSyncTime err:", err)
    return
  }
  lastTime := ee.Se.GetLastSyncTime()
  Debug(ee.Se, "lastTime", lastTime, "syncTime", syncTime)
  // os.Exit(1)
  if lastTime < syncTime {
    //Pull
    Normal(ee.Se, "Pull begin: ", ee.Name, ee.Se.Supname)
    ee.Se.RestCostTime()
    cli := NewFtpcli(ee.Se)
    cli.Pull()
  }

  //SyncNext
  go checkSyncNext(ee.Se)
}

func Push(ee EventEntry) {
  Debug(ee.Se, "Push Lock: ", ee.Name, ee.Se.Supname)

  lock := getLock("Push:"+ee.Se.Supname)
  lock.Lock()
	defer func() {
		lock.Unlock()
		Debug(ee.Se, "Unlocked: Push", ee.Name, ee.Se.Supname)
	}()
  Debug(ee.Se, "Push Locked: ", ee.Name, ee.Se.Supname)

  syncTime, err := ee.Se.GetCurrentSyncTime()
  if err != nil {
    Normal(ee.Se, "Pull se.GetCurrentSyncTime err:", err)
    return
  }
  lastTime := ee.Se.GetLastSyncTime()
  if lastTime < syncTime {
    //Pull
    Normal(ee.Se, "Push begin: ", ee.Name, ee.Se.Supname)
    ee.Se.RestCostTime()
    cli := NewFtpcli(ee.Se)
    cli.Push()
  }

  //SyncNext
  go checkSyncNext(ee.Se)
}

func checkSyncNext(se SyncEntry) {
  // Debug("checkSyncNext: ", se)
  if se.SyncNext == nil || len(se.SyncNext) == 0 {
    return
  }
  syncTime, err := se.GetCurrentSyncTime()
  if err != nil {
    Normal(se, "checkSyncNext se.GetCurrentSyncTime err:", err)
    return
  }
  lastTime := se.GetLastSyncTime()
  if lastTime >= syncTime {
    for _,nextse := range se.SyncNext {
      nextseSyncTime, err := nextse.GetCurrentSyncTime()
      if err != nil {
        Normal(se, "checkSyncNext nextse.GetCurrentSyncTime err:", err)
        return
      }
      nextseLastTime := nextse.GetLastSyncTime()
      if nextseLastTime < nextseSyncTime {
        Send(EventEntry{nextse.Method, nextse})
      } else {
        checkSyncNext(nextse)
      }
    }
  }
}

func Sh(ee EventEntry) {
  Debug(ee.Se, "Sh Lock: ", ee.Name, ee.Se.Supname)
  lock := getLock("Sh:"+ee.Se.Supname)
  lock.Lock()
	defer func() {
		lock.Unlock()
		Debug(ee.Se, "Unlocked: Sh", ee.Name, ee.Se.Supname)
	}()
  Debug(ee.Se, "Sh Locked: ", ee.Name, ee.Se.Supname)

  syncTime, err := ee.Se.GetCurrentSyncTime()
  if err != nil {
    Normal(ee.Se, "Sh se.GetCurrentSyncTime err:", err)
    return
  }
  lastTime := ee.Se.GetLastSyncTime()
  Debug(ee.Se, "lastTime", lastTime, "syncTime", syncTime)
  // os.Exit(1)
  if lastTime < syncTime {
    //Pull
    Normal(ee.Se, "Sh begin: ", ee.Name, ee.Se.Supname)
    ee.Se.RestCostTime()
    cli := NewCmdcli(ee.Se)
    cli.Sh()
  }

  //SyncNext
  go checkSyncNext(ee.Se)
}

func Rsync(ee EventEntry) {
  Debug(ee.Se, "Rsync Lock: ", ee.Name, ee.Se.Supname)
  lock := getLock("Rsync:"+ee.Se.Supname)
  lock.Lock()
	defer func() {
		lock.Unlock()
		Debug(ee.Se, "Unlocked: Rsync", ee.Name, ee.Se.Supname)
	}()
  Debug(ee.Se, "Rsync Locked: ", ee.Name, ee.Se.Supname)

  syncTime, err := ee.Se.GetCurrentSyncTime()
  if err != nil {
    Normal(ee.Se, "Rsync se.GetCurrentSyncTime err:", err)
    return
  }
  lastTime := ee.Se.GetLastSyncTime()
  Debug(ee.Se, "lastTime", lastTime, "syncTime", syncTime)
  // os.Exit(1)
  if lastTime < syncTime {
    //Pull
    Normal(ee.Se, "Rsync begin: ", ee.Name, ee.Se.Supname)
    ee.Se.RestCostTime()
    cli := NewRsynccli(ee.Se)
    cli.Rsync()
  }

  //SyncNext
  go checkSyncNext(ee.Se)
}

func Import(ee EventEntry) {
  Debug(ee.Se, "Import Lock: ", ee.Name, ee.Se.Supname)
  lock := getLock("Import:"+ee.Se.Supname)
  lock.Lock()
	defer func() {
		lock.Unlock()
		Debug(ee.Se, "Unlocked: Import", ee.Name, ee.Se.Supname)
	}()
  Debug(ee.Se, "Import Locked: ", ee.Name, ee.Se.Supname)

  syncTime, err := ee.Se.GetCurrentSyncTime()
  if err != nil {
    Normal(ee.Se, "Import se.GetCurrentSyncTime err:", err)
    return
  }
  lastTime := ee.Se.GetLastSyncTime()
  Debug(ee.Se, "lastTime", lastTime, "syncTime", syncTime)
  // os.Exit(1)
  if lastTime < syncTime {
    //Pull
    Normal(ee.Se, "Import begin: ", ee.Name, ee.Se.Supname)
    ee.Se.RestCostTime()
    cli := NewCmdcli(ee.Se)
    cli.Import()
  }

  //SyncNext
  go checkSyncNext(ee.Se)
}

var EventMap = map[string]Event{
  "pull": Pull,
  "push": Push,
  "sh": Sh,
  "rsync": Rsync,
  "import": Import,
}

func NewSyncEntry(cnf *config.Config, conf string, supname string) (SyncEntry, error) {
  se := SyncEntry{
    Conf: conf,
    Supname: supname,
    Code: "",
    LogPath: "./",
    SyncTime: "",
    SyncEndTime: "",
    LastSyncTime: 0,
    st: 0,
  }

  //开始时间
  se.st = time.Now().UnixNano()

  //LogPath
  if cnf.HasSection("log") && cnf.HasOption("log", "path") {
    se.LogPath, _ = cnf.String("log", "path")
  }
  se.LogPath = strings.TrimRight(se.LogPath, "/") + "/"

  //Method
  if cnf.HasSection(supname) && cnf.HasOption(supname, "method") {
    se.Method, _ = cnf.String(supname, "method")
  } else if cnf.HasSection("main") && cnf.HasOption("main", "method") {
    se.Method, _ = cnf.String("main", "method")
  } else {
    return se, errors.New("missing method");
  }

  //Cmd
  if cnf.HasSection(supname) && cnf.HasOption(supname, "code") {
    se.Code, _ = cnf.String(supname, "code")
  } else {
    se.Code = ""
  }

  //SyncTime
  if cnf.HasSection(supname) && cnf.HasOption(supname, "sync-time") {
    se.SyncTime, _ = cnf.String(supname, "sync-time")
  } else if cnf.HasSection("main") && cnf.HasOption("main", "sync-time") {
    se.SyncTime, _ = cnf.String("main", "sync-time")
  } else {
    se.SyncTime = "00:00:01"
  }

  //SyncEndTime
  if cnf.HasSection(supname) && cnf.HasOption(supname, "sync-end-time") {
    se.SyncEndTime, _ = cnf.String(supname, "sync-end-time")
  } else if cnf.HasSection("main") && cnf.HasOption("main", "sync-end-time") {
    se.SyncEndTime, _ = cnf.String("main", "sync-end-time")
  } else {
    se.SyncEndTime = "23:00:01"
  }

  //SyncNext
  if cnf.HasSection(supname) && cnf.HasOption(supname, "sync-next") {
    syncNext, _ := cnf.String(supname, "sync-next")
    syncNextList := strings.Split(syncNext, ",")
    se.SyncNext = make(map[string]SyncEntry)
    for _, val := range syncNextList {
      val = strings.TrimSpace(val)
      newse, err := NewSyncEntry(cnf, conf, val)
      if err != nil {
        return se, err;
      }
      se.SyncNext[val] = newse
    }
  }

  return se, nil
}

func (se *SyncEntry) RestCostTime() {
  se.st = time.Now().UnixNano()
}

func (se SyncEntry) GetCostTime() float64 {
  et := time.Now().UnixNano()
  cost := float64(et - se.st) / 1e9
  return cost
}


func (se SyncEntry) GetLastPullTime() int64 {
  lastPullTime, err := ioutil.ReadFile(se.LogPath + se.Supname + "_pull.log")
  if err != nil {
    Debug(se, "GetLastPullTime: ", err)
    return 0
  }
  tmp, err := time.ParseInLocation(timeLayout, string(lastPullTime), time.Local)
  if err != nil {
    Debug(se, "GetLastPullTime time.Parse: ", err)
    return 0
  }

  return tmp.Unix()
}

func (se SyncEntry) UpdateLastPullTime(lastPullTime string) {
  se.RestCostTime()
  Debug(se, "UpdateLastPullTime: ", se.Supname)
  err := ioutil.WriteFile(se.LogPath + se.Supname + "_pull.log", []byte(lastPullTime), 0775)
  if err != nil {
    Normal(se, "UpdateLastPullTime err:", err)
    return
  }
}

func (se SyncEntry) GetLastSyncTime() int64 {
  // Debug("GetLastSyncTime: ", se)
  lastSyncTime, err := ioutil.ReadFile(se.LogPath + se.Supname + ".log")
  if err != nil {
    Debug(se, "GetLastSyncTime: ", err)
    return 0
  }
  lastSyncTimeTmp, err := time.ParseInLocation(timeLayout, string(lastSyncTime), time.Local)
  if err != nil {
    Debug(se, "GetLastSyncTime time.Parse: ", err)
    return 0
  }

  return lastSyncTimeTmp.Unix()
}

func (se SyncEntry) UpdateLastSyncTime() {
  se.RestCostTime()
  Debug(se, "UpdateLastSyncTime: ", se.Supname)
  lastSyncTime := time.Now().Format(timeLayout)
  err := ioutil.WriteFile(se.LogPath + se.Supname + ".log", []byte(lastSyncTime), 0775)
  if err != nil {
    Normal(se, "UpdateLastSyncTime err:", err)
    return
  }
  // se.LastSyncTime = time.Now().Unix()
}

func (se *SyncEntry) GetYesterdaySyncTime() int64 {
  syncTime, _ := se.GetCurrentSyncTime()
  if syncTime > 0 {
    syncTime = syncTime - (24 * 3600)
  }
  // Debug(se, "GetYesterdaySyncTime: ", syncTime)
  return syncTime
}

func (se *SyncEntry) GetCurrentSyncTime() (int64, error) {
  if len(se.SyncTime) == 0 {
    return 0, nil
  }
  day := time.Now().Format("2006-01-02 ")
	syncTimeTmp, err := time.ParseInLocation(timeLayout, day + se.SyncTime, time.Local)
	if err != nil {
    return 0, err
	}
	syncTime := syncTimeTmp.Unix()
  // Debug(se, "GetCurrentSyncTime: ", day + se.SyncTime, syncTime)
  return syncTime, nil
}

func (se *SyncEntry) GetSyncEndTime() (int64, error) {
  if len(se.SyncEndTime) == 0 {
    return 0, nil
  }
  day := time.Now().Format("2006-01-02 ")
  syncTimeTmp, err := time.ParseInLocation(timeLayout, day + se.SyncEndTime, time.Local)
  if err != nil {
    return 0, err
  }
  syncEndTime := syncTimeTmp.Unix()
  // Debug(se, "GetCurrentSyncTime: ", day + se.SyncTime, syncTime)
  return syncEndTime, nil
}

func Send(ee EventEntry) {
  ee.Se.RestCostTime()
  Normal(ee.Se, "Send: ", ee.Name, ee.Se.Supname)
  chEvent <- ee
}

func Recv(timeout time.Duration) (err error) {
  defer func(){
    if errval := recover(); errval!=nil {
      errmsg := fmt.Sprintf("Recv err: %v", errval)
      err = errors.New(errmsg)
    }
  }()

  Debug(nil, "Recv: start..")
  err = nil
  t := make(chan bool, 1)
  go func() {
      time.Sleep(timeout)
      t <- true
  }()

  select {
  case ee := <-chEvent:
      Normal(ee.Se, "Recv:", ee.Name, ee.Se.Supname)
      fn, ok := EventMap[ee.Name]
      if ok != true {
        Warning(ee.Se, "EventMap err:", err)
      }

      //deel
      ee.Se.RestCostTime()
      go fn(ee)
      return
    case <-t:
      return
  }
}

func Listen() {
  Debug(nil, "Listen: start..")
  for {
    if err := Recv(30 * time.Second); err != nil {
      fmt.Println(err)
    }
  }
}

func Watch(chSync chan CheckEntry, timeout time.Duration) {
  Debug(nil, "Watch: start..")
  for {
    t := make(chan bool, 1)
    go func() {
        time.Sleep(timeout)
        t <- true
    }()

    select {
    case ce := <-chSync:
        time.Sleep(timeout)
        // Debug("Watch: ok", ce)
        go Check(chSync, ce)
      case <-t:
    }
  }
}

func Check(chSync chan CheckEntry, ce CheckEntry) {
  Debug(nil, "Check:", ce)
  defer func(){
    if err := recover(); err!=nil {
      Debug(nil,"Check err: ", err)
    }
		chSync <- ce
  }()

  supname := ce.Supname
  cnf, err := config.ReadDefault(ce.Conf)
	if err != nil {
		Debug(nil,"Check config.ReadDefault err:", err)
    return
	}

  se, err := NewSyncEntry(cnf, ce.Conf, supname)
  if err != nil {
		Debug(nil,"Check NewSyncEntry err:", err)
    return
	}

	for {
		now := time.Now().Unix()
    //时间判断
    syncTime, err := se.GetCurrentSyncTime()
    if err != nil {
      Debug(se, "Check se.GetCurrentSyncTime err:", err, supname)
      return
    }

    syncEndTime, err := se.GetSyncEndTime()
    if err != nil {
      Debug(se, "Check se.GetSyncEndTime err:", err, supname)
      return
    }

    //现在时间大于当天同步时间
		if now >= syncTime && now <= syncEndTime {
			lastTime := se.GetLastSyncTime()
      //上次同步时间小于当天同步时间
			if lastTime < syncTime {
        Send(EventEntry{se.Method, se})
        time.Sleep(600 * time.Second)
				break
			}

      //SyncNext
      checkSyncNext(se)

			Debug(se, "Check ok:", lastTime, now, supname)
      break
		}

		time.Sleep(60 * time.Second)
	}
}



func Normal(se interface{}, msg ...interface{}) {
  if se == nil {
    return
  }
  log := fmt.Sprintf("%v", msg)
  DefLogger.ReadyRsync(log, NORMAL, se)
	DefLogger.Normal(msg...)
}

func Warning(se interface{}, msg ...interface{}) {
  log := fmt.Sprintf("%v", msg)
  DefLogger.ReadyRsync(log, WARNING, se)
  DefLogger.Warning(msg...)
}

func Debug(se interface{}, msg ...interface{}) {
  if se == nil {
    return
  }
  // log := fmt.Sprintf("%v", msg)
  // ReadyRsync(log, DEBUG, se)
  DefLogger.Debug(msg...)
}
