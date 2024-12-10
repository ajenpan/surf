package lobby

type Banker struct {
}

type PropItem struct {
	PropId     int32
	PropValue  int64
	UpdateAtMs int64
}

func (b *Banker) UpdateUserProp(uid uint32, propid uint32, chgv int64) error {

	return nil
}

func (b *Banker) GetUserProp(uid uint32) ([]*PropItem, error) {
	ret := []*PropItem{}

	return ret, nil
}
