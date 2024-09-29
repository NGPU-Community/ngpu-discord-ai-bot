package db

import (
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/nGPU/bot/configure"
	"github.com/nGPU/bot/header"
	log4plus "github.com/nGPU/common/log4go"
)

type PaymentState int

const (
	PaymentFail        PaymentState = -1 //支付失败
	PaymentNewSession  PaymentState = 0  //创建支付
	PaymentSuccess     PaymentState = 1  //支付成功
	PaymentSubscribeID PaymentState = 2  //支付订单完成
	PaymentShow        PaymentState = 3  //支付显示
)

type PaymentDB struct {
	outChan chan bool
	mysqlDb *MysqlManager
	roots   *x509.CertPool
	rootPEM []byte
}

var gPaymentDB *PaymentDB

// 是否已经订阅
func (p *PaymentDB) IsSubscribe(discordid string, email string) (error, bool) {
	funName := "IsSubscribe"
	if !p.mysqlDb.IsConnect() {
		log4plus.Error("%s Db Not Connect ---->>>>", funName)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName)), false
	}
	sql := fmt.Sprintf(`select 	IFNULL(id,-1), IFNULL(subscriptionid,''), IFNULL(email,''), IFNULL(state,-1) from subscription where discordid='%s' limit 1;`, discordid)
	log4plus.Info("%s Query SQL=%s", funName, sql)
	rows, err := p.mysqlDb.Mysqldb.Query(sql)
	if err != nil {
		log4plus.Error("%s Query Failed Error=%s SQL=%s", funName, err.Error(), sql)
		return err, false
	}
	defer rows.Close()

	for rows.Next() {
		var tmpSubscriptionid, tmpEmail string
		var tmpId, tmpState int
		scanErr := rows.Scan(&tmpId, &tmpSubscriptionid, &tmpEmail, &tmpState)
		if scanErr != nil {
			log4plus.Error("%s Scan Error=%s", funName, scanErr.Error())
			continue
		}
		if tmpState == 1 {
			return nil, true
		} else if tmpState == int(PaymentFail) || tmpState == int(PaymentNewSession) {
			return nil, false
		}
	}
	return nil, false
}

// 创建订单
func (p *PaymentDB) CreateSession(discordId, eMail, sessionID, productId, priceId, customerid string, price float64) error {
	funName := "CreateSession"
	if !p.mysqlDb.IsConnect() {
		log4plus.Error("%s Db Not Connect ---->>>>", funName)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName))
	}
	sql := fmt.Sprintf(`select 	IFNULL(id,-1), IFNULL(sessionid,''), IFNULL(email,'') from subscription where discordId='%s' limit 1;`, discordId)
	log4plus.Info("%s Query SQL=%s", funName, sql)
	rows, err := p.mysqlDb.Mysqldb.Query(sql)
	if err != nil {
		log4plus.Error("%s Query Failed Error=%s SQL=%s", funName, err.Error(), sql)
		return err
	}
	defer rows.Close()

	var exist bool = false
	for rows.Next() {
		var tmpSubscribeID, tmpEMail string
		var tmpId int
		scanErr := rows.Scan(&tmpId, &tmpSubscribeID, &tmpEMail)
		if scanErr != nil {
			log4plus.Error("%s Scan Error=%s", funName, scanErr.Error())
			continue
		}
		log4plus.Info("%s tmpId=%d", funName, tmpId)
		if tmpId != int(PaymentFail) {
			exist = true
		}
	}

	log4plus.Info("%s exist=%t", funName, exist)
	if !exist {
		//Insert
		sql := fmt.Sprintf(`insert into subscription (sessionid, productid, priceid, customerid, email, discordid, price, updatetime, state) values ('%s', '%s', '%s', '%s', '%s', '%s', %.2f, NOW(), %d);`,
			sessionID, productId, priceId, customerid, eMail, discordId, price, int(PaymentNewSession))
		log4plus.Info("%s Insert SQL=%s", funName, sql)
		_, err := p.mysqlDb.Mysqldb.Exec(sql)
		if err != nil {
			log4plus.Error("%s No Rows err=%s userkey=%s", funName, err.Error(), discordId)
			return err
		}
		return nil
	} else {
		//Update
		sql := fmt.Sprintf(`update subscription set sessionid='%s', productid='%s', priceid='%s', customerid='%s', email='%s', price=%.2f, updatetime=NOW(), state=%d where discordid='%s';`,
			sessionID, productId, priceId, customerid, eMail, price, int(PaymentNewSession), discordId)
		log4plus.Info("%s Update SQL=%s", funName, sql)
		_, err := p.mysqlDb.Mysqldb.Exec(sql)
		if err != nil {
			log4plus.Error("%s No Rows err=%s userkey=%s", funName, err.Error(), discordId)
			return err
		}
		return nil
	}
}

// 更新订单成功
func (p *PaymentDB) PaymentSuccess(sessionID string, subscriptionId string) error {
	funName := "PaymentSuccess"
	if !p.mysqlDb.IsConnect() {
		log4plus.Error("%s Db Not Connect ---->>>>", funName)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName))
	}
	//Update
	sql := fmt.Sprintf(`update subscription set state=%d, subscriptionid='%s' where sessionid='%s';`, int(PaymentSuccess), subscriptionId, sessionID)
	log4plus.Info("%s Update SQL=%s", funName, sql)
	_, err := p.mysqlDb.Mysqldb.Exec(sql)
	if err != nil {
		log4plus.Error("%s No Rows err=%s sessionid=%s", funName, err.Error(), sessionID)
		return err
	}
	return nil
}

// 更新订单ID
func (p *PaymentDB) SubscribeID(sessionID string, subscriptionid string) error {
	funName := "SubscribeID"
	if !p.mysqlDb.IsConnect() {
		log4plus.Error("%s Db Not Connect ---->>>>", funName)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName))
	}
	//Update
	sql := fmt.Sprintf(`update subscription set state=%d, subscriptionid='%s' where sessionid='%s';`, int(PaymentSubscribeID), subscriptionid, sessionID)
	log4plus.Info("%s Update SQL=%s", funName, sql)
	_, err := p.mysqlDb.Mysqldb.Exec(sql)
	if err != nil {
		log4plus.Error("%s No Rows err=%s sessionid=%s", funName, err.Error(), sessionID)
		return err
	}
	return nil
}

// 更新订单失败
func (p *PaymentDB) SubscribeFail(sessionID string) error {
	funName := "SubscribeFail"
	if !p.mysqlDb.IsConnect() {
		log4plus.Error("%s Db Not Connect ---->>>>", funName)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName))
	}
	//Update
	sql := fmt.Sprintf(`update subscription set state=%d where sessionid='%s';`, int(PaymentFail), sessionID)
	log4plus.Info("%s Update SQL=%s", funName, sql)
	_, err := p.mysqlDb.Mysqldb.Exec(sql)
	if err != nil {
		log4plus.Error("%s No Rows err=%s sessionid=%s", funName, err.Error(), sessionID)
		return err
	}
	return nil
}

// 订单显示
func (p *PaymentDB) SubscribeShow(sessionID string) error {
	funName := "SubscribeShow"
	if !p.mysqlDb.IsConnect() {
		log4plus.Error("%s Db Not Connect ---->>>>", funName)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName))
	}
	//Update
	sql := fmt.Sprintf(`update subscription set state=%d where sessionid='%s';`, int(PaymentShow), sessionID)
	log4plus.Info("%s Update SQL=%s", funName, sql)
	_, err := p.mysqlDb.Mysqldb.Exec(sql)
	if err != nil {
		log4plus.Error("%s No Rows err=%s sessionid=%s", funName, err.Error(), sessionID)
		return err
	}
	return nil
}

// 根据订单得到key
func (p *PaymentDB) GetSession(sessionId string) (error, header.StripeInfo) {
	funName := "GetSession"
	if !p.mysqlDb.IsConnect() {
		log4plus.Error("%s Db Not Connect ---->>>>", funName)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName)), header.StripeInfo{}
	}
	sql := fmt.Sprintf(`select 	IFNULL(id,-1), 
								IFNULL(subscriptionid,''), 
								IFNULL(productid,''), 
								IFNULL(priceid,''), 
								IFNULL(customerid,''), 
								IFNULL(email,''), 
								IFNULL(userkey,''), 
								IFNULL(price,-1), 
								IFNULL(updatetime,''), 
								IFNULL(sessionid,''), 
								IFNULL(state,-1) from subscription where sessionid='%s' limit 1;`, sessionId)
	log4plus.Info("%s Query SQL=%s", funName, sql)
	rows, err := p.mysqlDb.Mysqldb.Query(sql)
	if err != nil {
		log4plus.Error("%s Query Failed Error=%s SQL=%s", funName, err.Error(), sql)
		return err, header.StripeInfo{}
	}
	defer rows.Close()

	var sessionInfo header.StripeInfo
	for rows.Next() {
		var tmpId int64
		scanErr := rows.Scan(&tmpId,
			&sessionInfo.Subscriptionid,
			&sessionInfo.Productid,
			&sessionInfo.Priceid,
			&sessionInfo.Customerid,
			&sessionInfo.Email,
			&sessionInfo.Userkey,
			&sessionInfo.Price,
			&sessionInfo.Updatetime,
			&sessionInfo.Sessionid,
			&sessionInfo.State)
		if scanErr != nil {
			log4plus.Error("%s Scan Error=%s", funName, scanErr.Error())
			continue
		}
		return nil, sessionInfo
	}
	return errors.New(fmt.Sprintf("%s not found sessionId", funName)), header.StripeInfo{}
}

func SingtonPaymentDB() *PaymentDB {
	if gPaymentDB == nil {
		log4plus.Info("SingtonPaymentDB ---->>>>")
		gPaymentDB = &PaymentDB{}
		if gPaymentDB.mysqlDb = NewMysql(configure.SingtonConfigure().Mysql.MysqlIp,
			configure.SingtonConfigure().Mysql.MysqlPort,
			configure.SingtonConfigure().Mysql.MysqlDBName,
			configure.SingtonConfigure().Mysql.MysqlDBCharset,
			configure.SingtonConfigure().Mysql.UserName,
			configure.SingtonConfigure().Mysql.Password); gPaymentDB.mysqlDb == nil {
			log4plus.Error("SingtonPaymentDB NewMysql Failed")
			return nil
		}
	}
	return gPaymentDB
}
