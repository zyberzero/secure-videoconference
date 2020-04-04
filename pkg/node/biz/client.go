package biz

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	nprotoo "github.com/cloudwebrtc/nats-protoo"
	"github.com/pion/ion/pkg/discovery"
	"github.com/pion/ion/pkg/log"
	"github.com/pion/ion/pkg/proto"
	"github.com/pion/ion/pkg/signal"
	"github.com/pion/ion/pkg/util"
)

// Entry is the biz entry
func Entry(method string, peer *signal.Peer, msg map[string]interface{}, accept signal.AcceptFunc, reject signal.RejectFunc) {
	log.Infof("method => %s, data => %v", method, msg)
	var result map[string]interface{}
	err := util.NewNpError(400, fmt.Sprintf("Unkown method [%s]", method))

	switch method {
	case proto.ClientClose:
		result, err = clientClose(peer, msg)
	case proto.ClientLogin:
		result, err = login(peer, msg)
	case proto.ClientJoin:
		result, err = join(peer, msg)
	case proto.ClientLeave:
		result, err = leave(peer, msg)
	case proto.ClientPublish:
		result, err = publish(peer, msg)
	case proto.ClientUnPublish:
		result, err = unpublish(peer, msg)
	case proto.ClientSubscribe:
		result, err = subscribe(peer, msg)
	case proto.ClientUnSubscribe:
		result, err = unsubscribe(peer, msg)
	case proto.ClientBroadcast:
		result, err = broadcast(peer, msg)
	case proto.ClientTrickleICE:
		result, err = trickle(peer, msg)
	}

	if err != nil {
		reject(err.Code, err.Reason)
	} else {
		accept(result)
	}
}

func getRPCForIslb() (*nprotoo.Requestor, bool) {
	for _, item := range services {
		if item.Info["service"] == "islb" {
			id := item.Info["id"]
			rpc, found := rpcs[id]
			if !found {
				rpcID := discovery.GetRPCChannel(item)
				log.Infof("Create rpc [%s] for islb", rpcID)
				rpc = protoo.NewRequestor(rpcID)
				rpcs[id] = rpc
			}
			return rpc, true
		}
	}
	log.Warnf("No islb node was found.")
	return nil, false
}

func handleSFUBroadCast(msg map[string]interface{}, subj string) {
	go func(msg map[string]interface{}) {
		method := util.Val(msg, "method")
		data := msg["data"].(map[string]interface{})
		log.Infof("handleSFUBroadCast: method=%s, data=%v", method, data)
		rid := util.Val(data, "rid")
		uid := util.Val(data, "uid")
		switch method {
		case proto.SFUTrickleICE:
			signal.NotifyAllWithoutID(rid, uid, proto.ClientOnStreamAdd, data)
		case proto.SFUStreamRemove:
			mid := util.Val(data, "mid")
			islb, found := getRPCForIslb()
			if found {
				islb.AsyncRequest(proto.IslbOnStreamRemove, util.Map("mid", mid))
			}
		}
	}(msg)
}

func getRPCForSFU(mid string) (string, *nprotoo.Requestor, *nprotoo.Error) {
	islb, found := getRPCForIslb()
	if !found {
		return "", nil, util.NewNpError(500, "Not found any node for islb.")
	}
	result, err := islb.SyncRequest(proto.IslbFindService, util.Map("service", "sfu", "mid", mid))
	if err != nil {
		return "", nil, err
	}

	log.Infof("SFU result => %v", result)
	rpcID := result["rpc-id"].(string)
	eventID := result["event-id"].(string)
	nodeID := result["id"].(string)
	rpc, found := rpcs[rpcID]
	if !found {
		rpc = protoo.NewRequestor(rpcID)
		protoo.OnBroadcast(eventID, handleSFUBroadCast)
		rpcs[rpcID] = rpc
	}
	return nodeID, rpc, nil
}

func login(peer *signal.Peer, msg map[string]interface{}) (map[string]interface{}, *nprotoo.Error) {
	log.Infof("biz.login peer.ID()=%s msg=%v", peer.ID(), msg)

	return emptyMap, nil
}

func loginBankID(peer *signal.Peer, msg map[string]interface{}) (map[string]interface{}, error) {
	// authenticate with bank ID

	delay := 60

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	grandAPI := os.Getenv("GRANDID_API")
	grandService := os.Getenv("GRANDID_SERVICE")
	var infoMap map[string]interface{}
	infoMap, _ = msg["info"].(map[string]interface{})
	name := util.Val(infoMap, "name")

	log.Infof("biz.loginBankID api=" + grandAPI + " service=" + grandService + " name=" + name + " rid=" + util.Val(msg, "rid"))

	url := "https://client.grandid.com/json1.1/FederatedLogin?apiKey=" + grandAPI + "&authenticateServiceKey=" + grandService
	bodyText := "thisDevice=false&mobileBankId=true&askForSSN=false&personalNumber=" + name + "&gui=false"
	bodyReader := strings.NewReader(bodyText)
	contenttype := "application/x-www-form-urlencoded"

	log.Infof("biz.loginBankID url=%s body=%s", url, bodyText)

	resp, err := client.Post(url, contenttype, bodyReader)

	log.Infof("biz.loginBankID resp=%s err=%s", resp, err)
	if err != nil {
		return msg, err
	}
	defer resp.Body.Close()

	// get the json struct
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return msg, err
	}
	var data map[string]string
	log.Infof("biz.loginBankID body=%s", body)
	err = json.Unmarshal([]byte(body), &data)
	log.Infof("biz.loginBankID data=%s", data)
	sessionID := data["sessionId"]

	i := 0
	for i < delay {
		i++
		resp, err := client.Get("https://client.grandid.com/json1.1/GetSession?apiKey=" + grandAPI + "&authenticateServiceKey=" + grandService + "&sessionId=" + sessionID)
		if err != nil {
			return msg, err
		}
		defer resp.Body.Close()

		// get the json struct
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return msg, err
		}
		var data map[string]interface{}
		log.Infof("biz.loginBankID GetSession body=%s", body)
		err = json.Unmarshal([]byte(body), &data)
		log.Infof("biz.loginBankID GetSession data=%s", data)

		if errorObj, ok := data["errorObject"]; ok {
			errorObject, _ := errorObj.(map[string]interface{})
			message, _ := errorObject["message"].(map[string]interface{})
			status := util.Val(message, "status")

			log.Infof("biz.loginBankID errorObj=%s errorObject=%s message=%s status=%s", errorObj, errorObject, message, status)

			if "pending" == status {
				// 2 second is minimum wait time between requests to the API
				time.Sleep(2 * time.Second)
			} else {
				return msg, errors.New("BankID authentication failed: " + status + " - " + util.Val(message, "hintCode"))
			}
		} else {
			// we have a successful auth!

			attributes := data["userAttributes"].(map[string]interface{})
			name := attributes["name"]

			log.Infof("biz.loginBankID !!!!!!!!!SUCCESSFUL AUTH!!!!!!!!  name=%s", name)

			// Replace the name we got in the msg
			infoMap["name"] = name
			msg["info"] = infoMap
			return msg, nil
		}

	}

	return msg, errors.New("BankID authentication timed out")
}

// join room
func join(peer *signal.Peer, msg map[string]interface{}) (map[string]interface{}, *nprotoo.Error) {
	log.Infof("biz.join peer.ID()=%s msg=%v", peer.ID(), msg)
	if ok, err := verifyData(msg, "rid"); !ok {
		return nil, err
	}

	// authenticate
	msg, authErr := loginBankID(peer, msg)
	if authErr != nil {
		log.Infof("Error when authenticating with BankID: " + authErr.Error())
		return nil, util.NewNpError(500, "Error when authenticating: "+authErr.Error())
	}

	rid := util.Val(msg, "rid")
	//already joined this room
	if signal.HasPeer(rid, peer) {
		return emptyMap, nil
	}
	signal.AddPeer(rid, peer)

	islb, found := getRPCForIslb()
	if !found {
		return nil, util.NewNpError(500, "Not found any node for islb.")
	}
	// Send join => islb
	info := util.Val(msg, "info")
	uid := peer.ID()
	islb.SyncRequest(proto.IslbClientOnJoin, util.Map("rid", rid, "uid", uid, "info", info))
	// Send getPubs => islb
	islb.AsyncRequest(proto.IslbGetPubs, util.Map("rid", rid, "uid", uid)).Then(
		func(result map[string]interface{}) {
			log.Infof("IslbGetPubs: result=%v", result)
			if result["pubs"] == nil {
				return
			}
			pubs := result["pubs"].([]interface{})
			for _, pub := range pubs {
				info := pub.(map[string]interface{})["info"].(string)
				mid := pub.(map[string]interface{})["mid"].(string)
				uid := pub.(map[string]interface{})["uid"].(string)
				rid := result["rid"].(string)
				tracks := pub.(map[string]interface{})["tracks"].(map[string]interface{})

				var infoObj map[string]interface{}
				err := json.Unmarshal([]byte(info), &infoObj)
				if err != nil {
					log.Errorf("json.Unmarshal: err=%v", err)
				}
				log.Infof("IslbGetPubs: mid=%v info=%v", mid, infoObj)
				// peer <=  range pubs(mid)
				if mid != "" {
					peer.Notify(proto.ClientOnStreamAdd, util.Map("rid", rid, "uid", uid, "mid", mid, "info", infoObj, "tracks", tracks))
				}
			}
		},
		func(err *nprotoo.Error) {

		})

	return emptyMap, nil
}

func leave(peer *signal.Peer, msg map[string]interface{}) (map[string]interface{}, *nprotoo.Error) {
	log.Infof("biz.leave peer.ID()=%s msg=%v", peer.ID(), msg)
	defer util.Recover("biz.leave")

	if ok, err := verifyData(msg, "rid"); !ok {
		return nil, err
	}

	rid := util.Val(msg, "rid")
	uid := peer.ID()

	islb, found := getRPCForIslb()
	if !found {
		return nil, util.NewNpError(500, "Not found any node for islb.")
	}

	islb.AsyncRequest(proto.IslbOnStreamRemove, util.Map("rid", rid, "uid", uid, "mid", ""))
	islb.SyncRequest(proto.IslbClientOnLeave, util.Map("rid", rid, "uid", uid))
	signal.DelPeer(rid, peer.ID())
	return emptyMap, nil
}

func clientClose(peer *signal.Peer, msg map[string]interface{}) (map[string]interface{}, *nprotoo.Error) {
	log.Infof("biz.close peer.ID()=%s msg=%v", peer.ID(), msg)
	return leave(peer, msg)
}

func publish(peer *signal.Peer, msg map[string]interface{}) (map[string]interface{}, *nprotoo.Error) {
	log.Infof("biz.publish peer.ID()=%s", peer.ID())

	nid, sfu, err := getRPCForSFU("")
	if err != nil {
		log.Warnf("Not found any sfu node, reject: %d => %s", err.Code, err.Reason)
		return nil, util.NewNpError(err.Code, err.Reason)
	}

	jsep := msg["jsep"].(map[string]interface{})
	options := msg["options"].(map[string]interface{})
	room := signal.GetRoomByPeer(peer.ID())
	if room == nil {
		return nil, util.NewNpError(codeRoomErr, codeStr(codeRoomErr))
	}

	uid := peer.ID()
	result, err := sfu.SyncRequest(proto.ClientPublish, util.Map("uid", uid, "jsep", jsep, "options", options))
	if err != nil {
		log.Warnf("reject: %d => %s", err.Code, err.Reason)
		return nil, util.NewNpError(err.Code, err.Reason)
	}

	log.Infof("publish: result => %v", result)
	mid := util.Val(result, "mid")
	rid := room.ID()
	tracks := result["tracks"]
	islb, found := getRPCForIslb()
	if !found {
		return nil, util.NewNpError(500, "Not found any node for islb.")
	}
	islb.AsyncRequest(proto.IslbOnStreamAdd, util.Map("rid", rid, "nid", nid, "uid", uid, "mid", mid, "tracks", tracks))
	return result, nil
}

// unpublish from app
func unpublish(peer *signal.Peer, msg map[string]interface{}) (map[string]interface{}, *nprotoo.Error) {
	log.Infof("signal.unpublish peer.ID()=%s msg=%v", peer.ID(), msg)

	mid := util.Val(msg, "mid")
	rid := util.Val(msg, "rid")
	uid := peer.ID()

	_, sfu, err := getRPCForSFU(mid)
	if err != nil {
		log.Warnf("Not found any sfu node, reject: %d => %s", err.Code, err.Reason)
		return nil, err
	}

	_, err = sfu.SyncRequest(proto.ClientUnPublish, util.Map("mid", mid))
	if err != nil {
		return nil, err
	}

	islb, found := getRPCForIslb()
	if !found {
		return nil, util.NewNpError(500, "Not found any node for islb.")
	}
	// if this mid is a webrtc pub
	// tell islb stream-remove, `rtc.DelPub(mid)` will be done when islb broadcast stream-remove
	islb.AsyncRequest(proto.IslbOnStreamRemove, util.Map("rid", rid, "uid", uid, "mid", mid))
	return emptyMap, nil
}

func subscribe(peer *signal.Peer, msg map[string]interface{}) (map[string]interface{}, *nprotoo.Error) {
	log.Infof("biz.subscribe peer.ID()=%s ", peer.ID())

	if ok, err := verifyData(msg, "jsep", "mid"); !ok {
		return nil, err
	}
	mid := util.Val(msg, "mid")
	nodeID, sfu, err := getRPCForSFU(mid)
	if err != nil {
		log.Warnf("Not found any sfu node, reject: %d => %s", err.Code, err.Reason)
		return nil, util.NewNpError(err.Code, err.Reason)
	}

	// TODO:
	if nodeID != "node for mid" {
		log.Warnf("Not the same node, need to enable sfu-sfu relay!")
	}

	room := signal.GetRoomByPeer(peer.ID())
	uid := peer.ID()
	rid := room.ID()

	jsep := msg["jsep"].(map[string]interface{})

	islb, found := getRPCForIslb()
	if !found {
		return nil, util.NewNpError(500, "Not found any node for islb.")
	}

	result, err := islb.SyncRequest(proto.IslbGetMediaInfo, util.Map("rid", rid, "mid", mid))
	if err != nil {
		log.Warnf("reject: %d => %s", err.Code, err.Reason)
		return nil, util.NewNpError(err.Code, err.Reason)
	}
	result, err = sfu.SyncRequest(proto.ClientSubscribe, util.Map("uid", uid, "mid", mid, "tracks", result["tracks"], "jsep", jsep))
	if err != nil {
		log.Warnf("reject: %d => %s", err.Code, err.Reason)
		return nil, util.NewNpError(err.Code, err.Reason)
	}

	log.Infof("subscribe: result => %v", result)
	return result, nil
}

func unsubscribe(peer *signal.Peer, msg map[string]interface{}) (map[string]interface{}, *nprotoo.Error) {
	log.Infof("biz.unsubscribe peer.ID()=%s msg=%v", peer.ID(), msg)

	if ok, err := verifyData(msg, "mid"); !ok {
		return nil, err
	}
	mid := util.Val(msg, "mid")

	_, sfu, err := getRPCForSFU(mid)
	if err != nil {
		log.Warnf("Not found any sfu node, reject: %d => %s", err.Code, err.Reason)
		return nil, util.NewNpError(err.Code, err.Reason)
	}

	result, err := sfu.SyncRequest(proto.ClientUnSubscribe, util.Map("mid", mid))
	if err != nil {
		log.Warnf("reject: %d => %s", err.Code, err.Reason)
		return nil, util.NewNpError(err.Code, err.Reason)
	}

	log.Infof("publish: result => %v", result)
	return result, nil
}

func broadcast(peer *signal.Peer, msg map[string]interface{}) (map[string]interface{}, *nprotoo.Error) {
	log.Infof("biz.unsubscribe peer.ID()=%s msg=%v", peer.ID(), msg)

	if ok, err := verifyData(msg, "rid", "uid", "info"); !ok {
		return nil, err
	}

	islb, found := getRPCForIslb()
	if !found {
		return nil, util.NewNpError(500, "Not found any node for islb.")
	}
	rid, uid, info := util.Val(msg, "rid"), util.Val(msg, "uid"), util.Val(msg, "info")
	islb.AsyncRequest(proto.IslbOnBroadcast, util.Map("rid", rid, "uid", uid, "info", info))
	return emptyMap, nil
}

func trickle(peer *signal.Peer, msg map[string]interface{}) (map[string]interface{}, *nprotoo.Error) {
	log.Infof("biz.trickle peer.ID()=%s msg=%v", peer.ID(), msg)

	mid := util.Val(msg, "mid")

	if ok, err := verifyData(msg, "rid", "uid", "info"); !ok {
		return nil, err
	}

	_, sfu, err := getRPCForSFU(mid)
	if err != nil {
		log.Warnf("Not found any sfu node, reject: %d => %s", err.Code, err.Reason)
		return nil, util.NewNpError(err.Code, err.Reason)
	}

	trickle := msg["trickle"].(map[string]interface{})

	sfu.AsyncRequest(proto.ClientTrickleICE, util.Map("mid", mid, "trickle", trickle))
	return emptyMap, nil
}
