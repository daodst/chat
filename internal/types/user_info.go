package types

type UserChainInfo struct {
	Localpart   string   `json:"localpart"`
	Servername  string   `json:"servername"`
	LimitMode   string   `json:"limit_mode"`
	TelNumbers  []int64  `json:"tel_numbers"`
	Blacklist   []string `json:"blacklist"`
	Whitelist   []string `json:"whitelist"`
	AddressBook []string `json:"address_book"`
	ChatFee     string   `json:"chat_fee"`
	MortgageFee string   `json:"mortgage_fee"`
}

type RelationInvite struct {
	Inviter    string `json:"inviter"`
	Invitee    string `json:"invitee"`
	PresentFee string `json:"present_fee"`
}
