package background

import (
	"github.com/iyear/biligo"
	"log"
)

type BClient struct {
	*biligo.BiliClient
}

func (b *BClient) NewBClient(dedeUserID string, dedeUserIDCkMd5 string, SESSDATA string, biliJCT string) (*BClient, error) {

	client, err := biligo.NewBiliClient(&biligo.BiliSetting{
		Auth: &biligo.CookieAuth{
			DedeUserID:      dedeUserID,
			DedeUserIDCkMd5: dedeUserIDCkMd5,
			SESSDATA:        SESSDATA,
			BiliJCT:         biliJCT,
		},
	})
	if err != nil {
		log.Fatalf("init BiliClient err : %+V", err)
		return nil, err
	}

	return &BClient{
		client,
	}, nil
}

type FavList struct {
	List []*FavDetail `json:"list"`
}
type FavDetail struct {
	Meta *biligo.FavDetail
}

func (b *BClient) GetMyFAVList() (*FavList, error) {
	favList, err := b.FavGetMy()
	if err != nil {
		return nil, err
	}
	res := &FavList{}
	for _, fav := range favList.List {
		favDetail, err := b.GetFAVInfo(fav.ID)
		if err != nil {
			continue
		}
		b.FavGetResDetail(fav.ID)

	}
	return res, nil
}
func (b *BClient) GetFAVInfo(mild int64) (*biligo.FavDetail, error) {
	fav, err := b.FavGetDetail(mild)
	if err != nil {
		return &biligo.FavDetail{}, err
	}
	return fav, nil
}
