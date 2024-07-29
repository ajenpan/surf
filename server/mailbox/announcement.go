package mailbox

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ajenpan/surf/core/log"
	proto "github.com/ajenpan/surf/msg/mailbox"
)

type announcement struct {
	resp   *proto.AnnouncementResponse
	rwlock sync.RWMutex

	effectAt time.Time
	expireAt time.Time
}

func (an *announcement) Init() error {
	an.resp = &proto.AnnouncementResponse{}
	raw, _ := os.ReadFile(announcementFilePath)
	if len(raw) > 2 {
		err := protojson.Unmarshal(raw, an.resp)
		if err != nil {
			return err
		}
		an.effectAt, _ = time.ParseInLocation(TimeLayout, an.resp.EffectAt, time.Local)
		an.expireAt, _ = time.ParseInLocation(TimeLayout, an.resp.ExpireAt, time.Local)

		log.Infof("announcement init success, title:%v, content:%v, effectAt:%s, expireAt:%s", an.resp.Title, an.resp.Content, an.effectAt, an.expireAt)
	}
	return nil
}

func (an *announcement) write(in *proto.PublishAnnouncementRequest) error {
	raw, err := protojson.MarshalOptions{Multiline: true, EmitUnpopulated: true}.Marshal(in)
	if err != nil {
		return err
	}
	an.rwlock.Lock()
	defer an.rwlock.Unlock()

	an.resp.Reset()
	err = protojson.Unmarshal(raw, an.resp)
	if err != nil {
		return err
	}

	//store to a file
	return os.WriteFile(announcementFilePath, []byte(raw), 0644)
}

func (an *announcement) read(out *proto.AnnouncementResponse, caller string) {
	an.rwlock.RLock()
	defer an.rwlock.RUnlock()

	n := time.Now()

	out.ExpectValid = an.resp.ExpectValid
	out.CurrentVaild = an.resp.ExpectValid && n.After(an.effectAt) && n.Before(an.expireAt)

	if out.CurrentVaild || caller == "admin" {
		out.Title = an.resp.Title
		out.Content = an.resp.Content
		out.EffectAt = an.resp.EffectAt
		out.ExpireAt = an.resp.ExpireAt
	}
}
func (h *Handler) PublishAnnouncement(ctx context.Context, in *proto.PublishAnnouncementRequest, out *proto.PublishAnnouncementResponse) error {
	effectAt, err := time.ParseInLocation(TimeLayout, in.EffectAt, time.Local)
	if err != nil {
		return err
	}
	expireAt, err := time.ParseInLocation(TimeLayout, in.ExpireAt, time.Local)
	if err != nil {
		return err
	}
	if effectAt.After(expireAt) {
		return fmt.Errorf("effectAt is after expireAt")
	}
	return h.ann.write(in)
}

func (h *Handler) Announcement(ctx context.Context, in *proto.AnnouncementRequest, out *proto.AnnouncementResponse) error {
	role, _ := GetCallerRoleFromCtx(ctx)
	h.ann.read(out, role)
	return nil
}
