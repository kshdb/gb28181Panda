package sipServer

import (
	"crypto/md5"
	"fmt"
	"gb28181Panda/config"
	"gb28181Panda/log"
	"gb28181Panda/model"
	"gb28181Panda/util"
	"github.com/ghettovoice/gosip/sip"
	"net/http"
	"strings"
)

const (
	DefaultAlgorithm = "MD5"
	WWWHeader        = "WWW-Authenticate"
	ExpiresHeader    = "Expires"
)

type Authorization struct {
	*sip.Authorization
}

func (a *Authorization) Verify(username, passwd string) bool {
	//1、将 username,realm,password 依次组合获取 1 个字符串，并用算法加密的到密文 r1
	//是否校验域值
	var s1 string
	if config.SipOp.CheckRealm == 1 {
		s1 = fmt.Sprintf("%s:%s:%s", username, config.SipOp.Realm, passwd) //检验域值，从当前配置读取域值
	} else {
		s1 = fmt.Sprintf("%s:%s:%s", username, a.Realm(), passwd) //不检验域值，从设备取域值
	}
	r1 := a.getDigest(s1)
	//2、将 method，即REGISTER ,uri 依次组合获取 1 个字符串，并对这个字符串使用算法 加密得到密文 r2
	s2 := fmt.Sprintf("REGISTER:%s", a.Uri())
	r2 := a.getDigest(s2)
	if r1 == "" || r2 == "" {
		log.Error("Authorization algorithm wrong")
		return false
	}
	//3、将密文 1，nonce 和密文 2 依次组合获取 1 个字符串，并对这个字符串使用算法加密，获得密文 r3，即Response
	s3 := fmt.Sprintf("%s:%s:%s", r1, a.Nonce(), r2)
	r3 := a.getDigest(s3)
	//4、计算服务端和客户端上报的是否相等
	isGet := r3 == a.Response()
	//log.Info("[Password-Verify-密码校验情况]", isGet)
	return isGet
}

func (a *Authorization) getDigest(raw string) string {
	switch a.Algorithm() {
	case "MD5":
		return fmt.Sprintf("%x", md5.Sum([]byte(raw)))
	default: //如果没有算法，默认使用MD5
		return fmt.Sprintf("%x", md5.Sum([]byte(raw)))
	}
}

// Register express等于0 相当于注销后设备下线
func Register(req sip.Request, tx sip.ServerTransaction) {
	idx := strings.Index(req.Source(), ":")
	from, _ := req.From()
	fromIp := req.Source()[:idx]
	fromPort := req.Source()[idx+1:]
	id := from.Address.User().String()
	transferFromLog(fromIp, fromPort)
	log.Info("Register-Request", req)
	//注销的代码
	h := req.GetHeaders(ExpiresHeader)
	if len(h) != 1 {
		log.Error("从头部获取expires失败", req)
		return
	} else {
		expires := h[0].(*sip.Expires)
		// 如果v=0，则代表该请求是注销请求
		if expires.Equals(new(sip.Expires)) {
			//expires值为0,该请求是注销请求
			log.Info("Register-Logout")
			fromRequest, ok := util.DeviceFromRequest(req)
			if !ok {
				return
			}
			device, _ := fromRequest.DeviceDetail()
			if device.DeviceId != "" {
				device = fromRequest
				device.Status = "OFF"
				log.Info("deviceInfo", device)
				//插入数据库的
				_ = device.DeviceUpdate()
			}
		} else {
			if len(id) != 20 {
				transferToLog(fromIp, fromPort)
				log.Infof("错误的国标id: %s", id)
				response := sip.NewResponseFromRequest("", req, http.StatusForbidden, "Forbidden", "")
				_ = tx.Respond(response)
				return
			}
			passAuth := false
			// 不需要密码情况
			if config.SipOp.Password == "" {
				passAuth = true
			} else {
				// 需要密码情况 设备第一次上报，返回401和加密算法
				if headers := req.GetHeaders("Authorization"); len(headers) > 0 {
					log.Info("Register-With-Authorization")
					authenticateHeader := headers[0].(*sip.GenericHeader)
					auth := &Authorization{sip.AuthFromValue(authenticateHeader.Contents)}
					// 有些摄像头没有配置用户名的地方，用户名就是摄像头自己的国标id
					var username string
					username = id
					// 设备第二次上报，校验
					if auth.Verify(username, config.SipOp.Password) {
						log.Info("Check-Username-Password-Success")
						passAuth = true
					} else {
						log.Info("Check-Username-Password-Error")
						transferToLog(fromIp, fromPort)
						response := sip.NewResponseFromRequest("", req, http.StatusForbidden, "Forbidden", "")
						log.Info(response)
						_ = tx.Respond(response)
						return
					}
				}
			}
			if passAuth {
				//认证账号和密码成功
				log.Info("Auth-Success")
				//从回复数据中解析device设备信息
				fromRequest, ok := util.DeviceFromRequest(req)
				if !ok {
					log.Error("fromRequest-Error")
					return
				}
				//先查询当前设备是否在数据库中
				device, _ := fromRequest.DeviceDetail()
				//不存在插入数据
				if device.DeviceId == "" {
					device = fromRequest
					device.CreatedAt = util.GetCurrenTimeNow()
					device.UpdatedAt = util.GetCurrenTimeNow()
					device.RegisterAt = util.GetCurrenTimeNow()
					device.KeepaliveAt = util.GetCurrenTimeNow()
					device.Expires = expires.Value()
					device.Status = "ON"
					device.Ip = fromIp
					device.Port = fromPort
					log.Info("Insert Database Device", device)
					//插入数据库的
					_ = device.DeviceAdd()
				} else {
					var deviceSql model.Device
					deviceSql.RegisterAt = util.GetCurrenTimeNow()
					deviceSql.KeepaliveAt = util.GetCurrenTimeNow()
					deviceSql.Status = "ON"
					deviceSql.Ip = fromIp
					deviceSql.Port = fromPort
					deviceSql.DeviceId = device.DeviceId
					//	存在，更新数据
					log.Info("Update Database Device", device)
					_ = deviceSql.DeviceUpdate()
				}

				resp := sip.NewResponseFromRequest("", req, http.StatusOK, "OK", "")
				to, _ := resp.To()
				resp.ReplaceHeaders("To", []sip.Header{&sip.ToHeader{Address: to.Address, Params: sip.NewParams().Add("tag", sip.String{Str: util.RandString(9)})}})
				resp.RemoveHeader("Allow")
				expires := sip.Expires(3600)
				resp.AppendHeader(&expires)
				resp.AppendHeader(&sip.GenericHeader{
					HeaderName: "Date",
					Contents:   util.GetCurrenTimeNowFormat(),
				})
				transferToLog(fromIp, fromPort)
				log.Info("Register-Success", resp)
				_ = tx.Respond(resp)
				//查询设备的信息和通道信息
				QueryDeviceSip(device)
				QueryChannelSip(device)
			} else {
				response := sip.NewResponseFromRequest("", req, http.StatusUnauthorized, "Unauthorized", "")
				auth := fmt.Sprintf(
					`Digest realm="%s",algorithm=%s,nonce="%s"`,
					config.SipOp.Realm,
					DefaultAlgorithm,
					util.RandString(32),
				)
				response.AppendHeader(&sip.GenericHeader{
					HeaderName: WWWHeader,
					Contents:   auth,
				})
				transferToLog(fromIp, fromPort)
				log.Info("Register-Back-With-[WWW-Authenticate]-Header", response)
				_ = tx.Respond(response)
			}
		}
	}
}
