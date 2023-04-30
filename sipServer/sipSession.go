package sipServer

import (
	"encoding/json"
	"fmt"
	"gb28181Panda/log"
	"gb28181Panda/storage"
	"github.com/pkg/errors"
)

// SipTX sip会话事务
type SipTX struct {
	DeviceId  string `json:"deviceId,omitempty"`
	ChannelId string `json:"channelId,omitempty"`
	SSRC      string `json:"SSRC,omitempty"`
	CallId    string `json:"callId,omitempty"`
	FromTag   string `json:"fromTag,omitempty"`
	ToTag     string `json:"toTag,omitempty"`
	ViaBranch string `json:"viaBranch,omitempty"`
}

// 保存sip事务信息
func saveStreamSession(deviceId string, channelId string, ssrc string, callId string, fromTag string, toTag string, viaBranch string) {
	tx := SipTX{
		DeviceId:  deviceId,
		ChannelId: channelId,
		SSRC:      ssrc,
		CallId:    callId,
		FromTag:   fromTag,
		ToTag:     toTag,
		ViaBranch: viaBranch,
	}

	key := fmt.Sprintf("%s:%s", "Media:Stream:Transaction", ssrc)
	b, _ := json.MarshalIndent(tx, "", "  ")
	storage.Set(key, b)
}

func GetTx(ssrc string) (SipTX, error) {
	key := fmt.Sprintf("%s:%s", "Media:Stream:Transaction", ssrc)
	j, err := storage.Get(key)
	if err != nil {
		log.Error(err)
		return SipTX{}, errors.WithMessage(err, "gei sip session tx fail")
	}
	var tx SipTX
	err = json.Unmarshal([]byte(j), &tx)
	if err != nil {
		return SipTX{}, errors.WithMessage(err, "unmarshal json data to struct fail")
	}

	return tx, nil

}
