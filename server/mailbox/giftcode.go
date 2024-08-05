package mailbox

import (
	"context"
	"fmt"

	"regexp"
	"strings"
	"time"

	proto "github.com/ajenpan/surf/msg/openproto/mailbox"
	gamedbMod "github.com/ajenpan/surf/server/mailbox/database/models"
)

var RegGiftCode = regexp.MustCompile(`^[a-zA-Z0-9]{8}$`)

func (h *Handler) GenerateGiftCode(ctx context.Context, in *proto.GenerateGiftCodeRequest, out *proto.GenerateGiftCodeResponse) error {

	if _, err := GetAdminUIDFromCtx(ctx); err != nil {
		return fmt.Errorf("you have no permission")
	}

	checkGiftType := func(gt string) bool {
		if gt == "SR" {
			return true
		} else if strings.HasPrefix(gt, "Prop") {
			return true
		}
		return false
	}
	if !checkGiftType(in.GiftType) {
		return fmt.Errorf("gift type: %s is invalid", in.GiftType)
	}

	var err error
	expireTime, err := time.ParseInLocation(TimeLayout, in.ExpireAt, time.Local)

	if err != nil {
		return fmt.Errorf("expire time: %s is invalid", in.ExpireAt)
	}

	if expireTime.Before(time.Now()) {
		return fmt.Errorf("expire time: %s is invalid, must after now", in.ExpireAt)
	}

	giftcode := RandStr(8)

	if !RegGiftCode.MatchString(giftcode) {
		return fmt.Errorf("code generate failed, the code is %s", giftcode)
	}

	record := &gamedbMod.GiftCode{
		Code:        giftcode,
		GiftType:    in.GiftType,
		GiftCount:   int(in.GiftCount),
		ValidCount:  int(in.MaxExchangeCount),
		RemainCount: int(in.MaxExchangeCount),
		ExpireAt:    expireTime,
		Status:      0,
	}

	err = h.WGameDB.Create(record).Error
	if err != nil {
		return err
	}

	out.Code = record.Code
	return nil
}

func (h *Handler) GiftCodeList(ctx context.Context, in *proto.GiftCodeListRequest, out *proto.GiftCodeListResponse) error {
	if _, err := GetAdminUIDFromCtx(ctx); err != nil {
		return fmt.Errorf("you have no permission")
	}

	in.Page = CheckListPage(in.Page)

	total := int64(0)
	if err := h.WGameDB.Debug().Model(&gamedbMod.GiftCode{}).Count(&total).Error; err != nil {
		return err
	}
	out.Total = int32(total)

	var list []*gamedbMod.GiftCode

	if err := h.WGameDB.Debug().Model(&gamedbMod.GiftCode{}).Limit(int(in.Page.PageSize)).Offset(int(in.Page.PageNum * in.Page.PageSize)).Order("id DESC").Find(&list).Error; err != nil {
		return err
	}

	for _, v := range list {
		out.List = append(out.List, &proto.GiftCodeListResponse_GiftCodeRecord{
			Code:                v.Code,
			GiftType:            v.GiftType,
			GiftCount:           int32(v.GiftCount),
			MaxExchangeCount:    int32(v.ValidCount),
			RemainExchangeCount: int32(v.RemainCount),
			ExpireAt:            v.ExpireAt.Format(TimeLayout),
			CreateAt:            v.CreateAt.Format(TimeLayout),
			Id:                  int32(v.ID),
			Status:              int32(v.Status),
		})
	}

	return nil
}

func (h *Handler) ExchangeGiftCode(ctx context.Context, in *proto.ExchangeGiftCodeRequest, out *proto.ExchangeGiftCodeResponse) error {
	user, err := GetUserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("unknown user")
	}

	in.Code = strings.Trim(in.Code, " ")

	if !RegGiftCode.MatchString(in.Code) {
		out.Flag = 3
		out.Msg = "invalid code"
		return nil
	}

	//cache the code info maybe batter ?

	record := &gamedbMod.GiftCode{}
	res := h.WGameDB.Debug().Model(record).Where("code = ?", in.Code).Find(record)
	if res.Error != nil {
		return res.Error
	}

	out.GiftCount = int32(record.GiftCount)

	if res.RowsAffected == 0 {
		out.Flag = 3
		out.Msg = "invalid code"
		return nil
	}

	if record.RemainCount <= 0 {
		out.Flag = 1
		out.Msg = "兑换码无使用次数"
		return nil
	}
	if record.ExpireAt.Before(time.Now()) {
		out.Flag = 2
		out.Msg = "兑换码已经过期"
		return nil
	}
	if record.Status != 0 {
		out.Flag = 3
		out.Msg = "invalid code"
		return nil
	}
	var tempid int
	res = h.WGameDB.Raw("select id from `gift_exchange` where uid=? and gift_code=?", user.UID, in.Code).Scan(&tempid)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 {
		out.Flag = 4
		out.Msg = "您已经领取过该兑换码"
		return nil
	}

	res = h.WGameDB.Exec("update gift_code set remain_count = remain_count - 1 WHERE code = ? and remain_count > 0", in.Code)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		out.Flag = 1
		out.Msg = "兑换码无使用次数"
		return nil
	}

	res = h.WGameDB.Exec("INSERT INTO `gift_exchange` (`uid`, `gift_code`, `gift_stat`) VALUES (?, ?, ?, 1)", user.UID, in.Code)
	if res.RowsAffected != 1 {
		out.Flag = 4
		out.Msg = "您已经领取过该兑换码"
		return nil
	}
	//TODO: add award

	out.Msg = "ok"
	return nil
}

func (h *Handler) UpdateGiftCode(ctx context.Context, in *proto.UpdateGiftCodeRequest, out *proto.UpdateGiftCodeResponse) error {
	_, err := GetAdminUIDFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("you have no permission")
	}
	record := &gamedbMod.GiftCode{
		ID:     int(in.Id),
		Status: int8(in.Status),
	}
	res := h.WGameDB.Debug().Model(record).Select(gamedbMod.GiftCodeColumns.Status).Updates(record)
	if res.Error != nil {
		return res.Error
	}
	return nil
}
