package new_db

import (
	"context"
	"errors"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	"github.com/matrix-org/dendrite/new_feature"
	"github.com/matrix-org/util"
	"github.com/tendermint/tendermint/libs/strings"
)

var Db *xorm.Engine

func InitDb(conn string) error {
	ctx := context.Background()
	log := util.GetLogger(ctx)
	log.Infof("init db with conn: %s", conn)
	db, err := xorm.NewEngine("postgres", conn)
	if err != nil {
		log.WithError(err).Error("NewEngine postgres")
		return err
	}
	//2.sql
	db.ShowSQL(true)
	Db = db
	return nil
}

type CanInviteRes struct {
	Localpart   string      `json:"localpart"`
	LimitMode   string      `json:"limit_mode"`
	ChatFee     string      `json:"chat_fee"`
	MortgageFee string      `json:"mortgage_fee"`
	Servername  string      `json:"servername"`
	Blacklist   StringArray `json:"blacklist"`
	Whitelist   StringArray `json:"whitelist"`
	PayedFee    string      `json:"payed_fee"`
}

// QueryAvailableListByLocals 
func QueryAvailableListByLocals(userLocal string, locals []string) (res []CanInviteRes, err error) {

	err = Db.Table("account_chain_data").Alias("a").Join("LEFT", []string{"account_relation_invite", "b"}, "a.localpart=b.invitee AND b.inviter=?", userLocal).Cols(
		"a.localpart,a.limit_mode,a.chat_fee,a.mortgage_fee,a.servername,a.blacklist,a.whitelist",
	).
		Where("(a.limit_mode='any' OR ?=ANY(a.whitelist) OR b.present_fee IS NOT NULL) AND ? <> ALL(COALESCE(a.blacklist, '{}'))", userLocal, userLocal).In("a.localpart", locals).
		Find(&res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// QueryNeedPayByLocals 
func QueryNeedPayByLocals(userLocal string, locals []string) (res []CanInviteRes, err error) {
	err = Db.Table("account_chain_data").Alias("a").Join("LEFT", []string{"account_relation_invite", "b"}, "a.localpart=b.invitee AND b.inviter=?", userLocal).Cols(
		"a.localpart,a.limit_mode,a.chat_fee,a.mortgage_fee,a.servername,a.blacklist,a.whitelist",
	).
		Where("a.limit_mode='fee' AND b.present_fee IS NULL AND ? <> ALL(COALESCE(a.blacklist, '{}')) AND ? <> ALL(COALESCE(a.whitelist, '{}'))", userLocal, userLocal).In("a.localpart", locals).
		Find(&res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type UserRes struct {
	DisplayName string `json:"display_name"` 
	AvatarURL   string `json:"avatar_url"`   

	Localpart   string `json:"localpart"`    
	Servername  string `json:"servername"`   
	LimitMode   string `json:"limit_mode"`   //  any fee list
	ChatFee     string `json:"chat_fee"`     
	MortgageFee string `json:"mortgage_fee"` 

	TelNumbers Int64Array  `json:"tel_numbers"` 
	Blacklist  StringArray `json:"blacklist"`   
	Whitelist  StringArray `json:"whitelist"`   

	CanWeTalk  bool   `json:"can_we_talk"`  
	CanPayTalk bool   `json:"can_pay_talk"` 
	PayedFee   string `json:"payed_fee"`    
}

// GetUserByPhone 
func GetUserByPhone(userLocal string, phone int64) (res UserRes, err error) {
	res = UserRes{}
	get, err := Db.SQL(`SELECT "a"."localpart", "a"."limit_mode", "a"."chat_fee", "a"."mortgage_fee", "a"."servername", "a"."blacklist", "a"."whitelist", (COALESCE(b.present_fee, '')) as payed_fee, "a"."tel_numbers" FROM "account_chain_data" AS "a" LEFT JOIN "account_relation_invite" AS "b" ON a.localpart=b.invitee AND b.inviter=? WHERE (? = ANY(a.tel_numbers)) LIMIT 1`, userLocal, phone).Get(&res)
	if err != nil {
		return UserRes{}, err
	}
	if !get {
		return UserRes{}, errors.New("user not found")
	}

	switch res.LimitMode {
	case "any":
		res.CanWeTalk = true
	case "fee":
		if res.PayedFee != "" || new_feature.JudgeIfPayByLocals(userLocal, res.Localpart) {
			res.CanWeTalk = true
		} else {
			res.CanPayTalk = true
		}
	}
	if strings.StringInSlice(userLocal, res.Whitelist) {
		res.CanWeTalk = true
	}
	if strings.StringInSlice(userLocal, res.Blacklist) {
		res.CanWeTalk = false
		res.CanPayTalk = false
	}
	return
}
