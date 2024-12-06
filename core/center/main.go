package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	etcclientv3 "go.etcd.io/etcd/client/v3"
)

var log = slog.Default()

type WatchPrefixs struct {
	NodeRegKey string `json:"node_reg_key"`
}
type Config struct {
	LogLevel       string             `json:"log_level"` // options: "panic|fatal|error|warn|info|debug|trace"
	EtcdConf       etcclientv3.Config `json:"etcd_conf"`
	WatchPrefixs   WatchPrefixs       `json:"watch_prefixs"`
	IDPoolStartAt  int                `json:"idpool_startat"`
	IDPoolEndAt    int                `json:"idpool_endat"`
	HttpListenPort int                `json:"http_listen_port"`
}

var DefaultConf = Config{
	LogLevel:       "info",
	IDPoolStartAt:  10001,
	IDPoolEndAt:    99999,
	HttpListenPort: 18080,
	WatchPrefixs: WatchPrefixs{
		NodeRegKey: "/nodes",
	},
	EtcdConf: etcclientv3.Config{
		Endpoints:   []string{"192.168.135.40:12379", "192.168.135.40:22379", "192.168.135.40:22379"},
		DialTimeout: 15 * time.Second,
	},
}

type IDInfoStatus = int32

const (
	IDInfoStatus_Unused IDInfoStatus = iota
	IDInfoStatus_Applying
	IDInfoStatus_Used
)

type IDInfo struct {
	Appid       int    `json:"appid"`
	Stat        int32  `json:"status"`
	EtcdKey     string `json:"etcd_key"`
	Ip          string `json:"remote_ip,omitempty"`
	Localip     string `json:"local_ip,omitempty"`
	ContainerId string `json:"container_id,omitempty"`
	AllocAt     string `json:"-"`
	OnlineAt    string `json:"-"`
	rwlock      sync.RWMutex
}

func (info *IDInfo) Clone() *IDInfo {
	info.rwlock.RLock()
	defer info.rwlock.RUnlock()
	ret := &IDInfo{
		Appid:       info.Appid,
		Stat:        info.Stat,
		EtcdKey:     info.EtcdKey,
		Ip:          info.Ip,
		ContainerId: info.ContainerId,
		AllocAt:     info.AllocAt,
		OnlineAt:    info.OnlineAt,
	}
	return ret
}

func (info *IDInfo) Reset() {
	info.rwlock.Lock()
	defer info.rwlock.Unlock()

	atomic.StoreInt32(&info.Stat, IDInfoStatus_Unused)

	info.Appid = 0
	info.EtcdKey = ""
	info.Ip = ""
	info.ContainerId = ""
	info.AllocAt = ""
	info.OnlineAt = ""
}

var RWLock sync.RWMutex
var UsedID = make(map[int]*IDInfo)
var UnusedIDList []*IDInfo

func SetUsed(appid []int) {
	RWLock.Lock()
	defer RWLock.Unlock()

	for _, v := range appid {
		UsedID[v] = &IDInfo{
			Appid: v,
			Stat:  IDInfoStatus_Used,
		}
	}
}

func GenAppid() {
	startAt := DefaultConf.IDPoolStartAt
	endAt := DefaultConf.IDPoolEndAt
	poolsize := endAt - startAt + 1
	if poolsize <= 0 {
		panic(fmt.Sprintf("poolsize <= %v", poolsize))
	}

	RWLock.Lock()
	defer RWLock.Unlock()

	UnusedIDList = make([]*IDInfo, 0, poolsize)

	for i := startAt; i < endAt; i++ {
		_, has := UsedID[int(i)]
		if has {
			continue
		}
		UnusedIDList = append(UnusedIDList, &IDInfo{Appid: i})
	}
}

func GetIDInfoByAppid(appid int) *IDInfo {
	RWLock.Lock()
	defer RWLock.Unlock()
	info, has := UsedID[appid]
	if has {
		return info
	}
	ret := &IDInfo{Appid: appid, Stat: IDInfoStatus_Unused}
	UsedID[ret.Appid] = ret
	return ret
}

func WaitPutback(info *IDInfo) {
	ok := atomic.CompareAndSwapInt32(&info.Stat, IDInfoStatus_Unused, IDInfoStatus_Applying)
	if !ok {
		log.Warn("WaitPutback stat err")
		return
	}

	const LockTime = 45 * time.Second
	time.AfterFunc(LockTime, func() {
		chg := atomic.LoadInt32(&info.Stat)
		if chg == IDInfoStatus_Applying {
			PutAppidBack(info.Appid)
		}
	})
}

func IsInMyRange(appid int) bool {
	return appid >= DefaultConf.IDPoolStartAt && appid < DefaultConf.IDPoolEndAt
}

func ApplyOne() *IDInfo {
	RWLock.Lock()
	defer RWLock.Unlock()
	e := len(UnusedIDList)
	if e == 0 {
		return nil
	}
	var ret *IDInfo
	cutAt := 0

	for i, v := range UnusedIDList {
		cutAt = i
		if _, has := UsedID[v.Appid]; !has {
			ret = v
			break
		}
	}

	UnusedIDList = UnusedIDList[cutAt+1:]
	if ret == nil {
		return nil
	}

	UsedID[ret.Appid] = ret

	WaitPutback(ret)
	return ret
}

func PutAppidBack(appid int) {
	if !IsInMyRange(appid) {
		log.Warn("PutAppidBack appid is not inrange", "appid", appid)
		return
	}

	RWLock.Lock()
	defer RWLock.Unlock()

	log.Info("PutAppidBack", "appid", appid, "unusedlen", len(UnusedIDList))

	info, has := UsedID[appid]
	if !has {
		return
	}

	delete(UsedID, appid)
	info.Reset()

	UnusedIDList = append(UnusedIDList, info)
}

func SetAppidStatUsed(appid int, etcdkey string) {
	if !IsInMyRange(appid) {
		log.Warn("SetAppidStatUsed appid is not inrange", "appid", appid)
		return
	}

	RWLock.Lock()
	defer RWLock.Unlock()

	info, has := UsedID[appid]
	if has {
		if atomic.LoadInt32(&info.Stat) == IDInfoStatus_Used {
			return
		}
		atomic.StoreInt32(&info.Stat, IDInfoStatus_Used)

		info.rwlock.Lock()
		defer info.rwlock.Unlock()
		info.EtcdKey = etcdkey
		info.OnlineAt = time.Now().Format(time.DateTime)
		return
	}

	UsedID[appid] = &IDInfo{
		Appid:    appid,
		Stat:     IDInfoStatus_Used,
		EtcdKey:  etcdkey,
		OnlineAt: time.Now().Format(time.DateTime),
		AllocAt:  time.Now().Format(time.DateTime),
	}
}

func GetUsedAppid() []*IDInfo {
	RWLock.Lock()
	defer RWLock.Unlock()
	ret := make([]*IDInfo, 0, len(UsedID))
	for _, v := range UsedID {
		ret = append(ret, v.Clone())
	}
	return ret
}

func ListUsedAppid(w http.ResponseWriter, r *http.Request) {
	resp := &struct {
		List []*IDInfo `json:"list"`
	}{
		List: GetUsedAppid(),
	}
	respRaw, _ := json.MarshalIndent(resp, "", " ")
	w.Write(respRaw)
	w.Header().Set("Content-Type", "application/json")
}

func CutRemoteAddr(addr string) (ip string) {
	idx := strings.LastIndex(addr, ":")
	return addr[0:idx]
}

func OnLogin(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	nodeidstr := values.Get("nodeid")
	passwd := values.Get("passwd")

	log.Info("OnLogin", "nodeid", nodeidstr, "passwd", passwd)
	resp := &struct {
		ErrMsg  string `json:"errmsg"`
		ErrCode int    `json:"errcode"`
		Appid   int    `json:"appid"`
		Ip      string `json:"ip"`
	}{}

	nodeid := 0
	if len(nodeidstr) > 0 {
		nodeid, _ = strconv.Atoi(nodeidstr)
	}
	remoteIp := CutRemoteAddr(r.RemoteAddr)

	var info *IDInfo

	if nodeid == 0 {
		if info = ApplyOne(); info != nil {
			info.rwlock.Lock()
			info.Ip = remoteIp
			info.ContainerId = passwd
			info.rwlock.Unlock()

			resp.Appid = info.Appid
		} else {
			resp.ErrCode = 2
			resp.ErrMsg = "id-pool is used out"
		}
	} else {
		info = GetIDInfoByAppid(nodeid)

		info.rwlock.Lock()
		stat := atomic.LoadInt32(&info.Stat)

		log.Debug("GetIDInfoByAppid", "nodeid", nodeid, "stat", stat, "infocontainerid", info.ContainerId, "containerid", passwd)

		switch stat {
		case IDInfoStatus_Unused:
			resp.Appid = info.Appid
			info.Ip = remoteIp
			info.ContainerId = passwd
			WaitPutback(info)
		case IDInfoStatus_Applying:
			fallthrough
		case IDInfoStatus_Used:
			if info.ContainerId == passwd {
				resp.Appid = info.Appid
			}
		}
		info.rwlock.Unlock()
		if resp.Appid == 0 {
			resp.ErrCode = 1
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func StartHttp() {
	http.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(CutRemoteAddr(r.RemoteAddr)))
	})

	http.HandleFunc("/login", OnLogin)
	http.HandleFunc("/used", ListUsedAppid)

	addr := fmt.Sprintf(":%d", DefaultConf.HttpListenPort)

	log.Info("http ListenAndServe", "addr", addr)
	http.ListenAndServe(addr, http.DefaultServeMux)
}

func InitConfig() {
	raw, err := os.ReadFile("config.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(raw, &DefaultConf)
	if err != nil {
		log.Error("InitConfig", "err", err)
	}
	if DefaultConf.IDPoolStartAt >= DefaultConf.IDPoolEndAt {
		panic("IDPoolStartAt >= IDPoolEndAt")
	}
}

func GetNodeIdFromKeystr(key string) int {
	list := strings.Split(key, "/")
	if len(list) < 1 {
		log.Warn("GetNodeIdFromKeystr", "key", key)
		return 0
	}
	nodeid, _ := strconv.Atoi(list[len(list)-1])
	return nodeid
}

func WatchServices(etcdcli *etcclientv3.Client, key string) error {
	watchKey := etcdcli.Watch(context.Background(), key, etcclientv3.WithPrefix())
	go func() {
		for resp := range watchKey {
			for _, event := range resp.Events {
				strkey := string(event.Kv.Key)

				jstr, _ := json.Marshal(event)

				log.Info("event", "type", event.Type, "key", strkey, "iscreated", event.IsCreate(), "jstr", string(jstr))

				appid := GetNodeIdFromKeystr(strkey)
				switch event.Type {
				case etcclientv3.EventTypeDelete:
					PutAppidBack(appid)
				case etcclientv3.EventTypePut:
					if event.IsCreate() {
						SetAppidStatUsed(appid, strkey)
					}
				}
			}
		}
	}()

	resp, err := etcdcli.Get(context.Background(), key, etcclientv3.WithKeysOnly(), etcclientv3.WithPrefix())
	if err != nil {
		return err
	}

	RWLock.Lock()
	defer RWLock.Unlock()

	for _, kv := range resp.Kvs {
		log.Info("getall", "key", string(kv.Key))
		appid := GetNodeIdFromKeystr(string(kv.Key))

		if appid != 0 {
			UsedID[appid] = &IDInfo{
				Appid:    appid,
				Stat:     IDInfoStatus_Used,
				EtcdKey:  string(kv.Key),
				OnlineAt: time.Now().Format(time.DateTime),
				AllocAt:  time.Now().Format(time.DateTime),
			}
		}
	}
	return nil
}

func WaitShutdown() os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	return <-signals
}

func RealMain(ctx *cli.Context) error {
	etcdcli, err := etcclientv3.New(DefaultConf.EtcdConf)
	if err != nil {
		return err
	}
	defer etcdcli.Close()

	if err := WatchServices(etcdcli, DefaultConf.WatchPrefixs.NodeRegKey); err != nil {
		return err
	}

	GenAppid()

	go StartHttp()

	log.Info("WaitShutdown..")
	WaitShutdown()
	return nil
}

func main() {
	InitConfig()

	app := cli.NewApp()
	app.Usage = "center for surf"
	app.Version = "1.0.0"
	app.Action = RealMain
	app.Commands = []*cli.Command{
		{
			Name:   "printconf",
			Hidden: true,
			Action: func(c *cli.Context) error {
				raw, err := json.Marshal(&DefaultConf)
				if err != nil {
					log.Error("CmdPrintConfig", "err", err)
					return err
				}
				fmt.Println(string(raw))
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Error("main", "err", err)
		os.Exit(-1)
	}
}
