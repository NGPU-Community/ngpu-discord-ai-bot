package db

import (
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"github.com/nGPU/bot/configure"
	"github.com/nGPU/bot/header"
	log4plus "github.com/nGPU/common/log4go"
)

type UserDB struct {
	outChan chan bool
	mysqlDb *MysqlManager
	roots   *x509.CertPool
	rootPEM []byte
}

var gUserDB *UserDB

// 创建用户
func (p *UserDB) CreateUser(discordId, eMail string) error {
	funName := "CreateUser"
	if !p.mysqlDb.IsConnect() {
		log4plus.Error("%s Db Not Connect ---->>>>", funName)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName))
	}
	userkey := fmt.Sprintf("%s%s", time.Now().Format("20060102150405"), fmt.Sprintf("%06d", time.Now().Nanosecond()/1e3))
	sql := fmt.Sprintf(`select IFNULL(id,-1) from user where discordId='%s' limit 1;`, discordId)
	log4plus.Info("%s Query SQL=[%s]", funName, sql)
	rows, err := p.mysqlDb.Mysqldb.Query(sql)
	if err != nil {
		log4plus.Error("%s Query Failed Error=[%s] SQL=[%s]", funName, err.Error(), sql)
		return err
	}
	defer rows.Close()

	var exist bool = false
	for rows.Next() {
		var tmpId int
		scanErr := rows.Scan(&tmpId)
		if scanErr != nil {
			log4plus.Error("%s Scan Error=%s", funName, scanErr.Error())
			continue
		}
		log4plus.Info("%s rows.Scan tmpId=[%d]", funName, tmpId)
		if tmpId != -1 {
			exist = true
		}
	}

	if !exist {
		//Insert
		sql := fmt.Sprintf(`insert into user (discordid, apiKey, email, remainingtime, subscribed, createtime, lasttime, state) values ('%s', '%s', '%s', 0, 0, NOW(), NOW(), 1);`, discordId, userkey, eMail)
		log4plus.Info("%s Insert SQL=[%s]", funName, sql)
		_, err := p.mysqlDb.Mysqldb.Exec(sql)
		if err != nil {
			log4plus.Error("%s No Rows err=%s apiKey=%s", funName, err.Error(), userkey)
			return err
		}
		return nil
	} else {
		errString := fmt.Sprintf("%s Failed to create user, discordId already exists discordId=[%s]", funName, discordId)
		log4plus.Error(errString)
		return errors.New(errString)
	}
}

// 查询用户信息
func (p *UserDB) GetUser(discordId string) (error, header.ResponseUserInfo) {
	funName := "GetUser"
	if !p.mysqlDb.IsConnect() {
		errString := fmt.Sprintf("%s Db Not Connect ---->>>>", funName)
		log4plus.Error(errString)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName)), header.ResponseUserInfo{
			ResultCode: 200,
			Msg:        errString,
		}
	}
	sql := fmt.Sprintf(`select 	IFNULL(id,-1), IFNULL(discordid,''), IFNULL(apiKey,''), IFNULL(email,''), IFNULL(remainingtime,0), IFNULL(subscribed,0), IFNULL(createtime,''), IFNULL(lasttime,''), IFNULL(state,-1) from user where discordid='%s' limit 1;`, discordId)
	log4plus.Info("%s Query SQL=[%s]", funName, sql)
	rows, err := p.mysqlDb.Mysqldb.Query(sql)
	if err != nil {
		errString := fmt.Sprintf("%s Query Failed Error=[%s] SQL=[%s]", funName, err.Error(), sql)
		log4plus.Error(errString)
		return err, header.ResponseUserInfo{
			ResultCode: 200,
			Msg:        errString,
		}
	}
	defer rows.Close()

	var userInfo header.ResponseUserInfo
	for rows.Next() {
		var tmpId int64
		scanErr := rows.Scan(&tmpId,
			&userInfo.DiscordId,
			&userInfo.UserKey,
			&userInfo.EMail,
			&userInfo.RemainingTime,
			&userInfo.Subscribed,
			&userInfo.CreateTime,
			&userInfo.Lasttime,
			&userInfo.State)
		if scanErr != nil {
			log4plus.Error("%s Scan Error=[%s]", funName, scanErr.Error())
			continue
		}
		userInfo.ResultCode = 200
		userInfo.Msg = "success"
		return nil, userInfo
	}
	errString := fmt.Sprintf("%s Query Failed SQL=[%s]", funName, sql)
	return errors.New(errString), header.ResponseUserInfo{
		ResultCode: 200,
		Msg:        errString,
	}
}

func (p *UserDB) CheckApiKey(apiKey string) (error, int) {
	funName := "CheckApiKey"
	if !p.mysqlDb.IsConnect() {
		errString := fmt.Sprintf("%s Db Not Connect ---->>>>", funName)
		log4plus.Error(errString)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName)), 0
	}
	sql := fmt.Sprintf(`select IFNULL(id,-1), subscribed, IFNULL(createtime,''), IFNULL(lasttime,''), IFNULL(state,0) from user where apiKey='%s';`, apiKey)
	log4plus.Info("%s Query SQL=[%s]", funName, sql)
	rows, err := p.mysqlDb.Mysqldb.Query(sql)
	if err != nil {
		errString := fmt.Sprintf("%s Query Failed Error=[%s] SQL=[%s]", funName, err.Error(), sql)
		log4plus.Error(errString)
		return err, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id, state int64
		var subscribed bool = false
		var createtime, lasttime string
		scanErr := rows.Scan(&id, &subscribed, &createtime, &lasttime, &state)
		if scanErr != nil {
			log4plus.Error("%s Scan Error=[%s]", funName, scanErr.Error())
			continue
		}
		if id == -1 {
			return nil, 0
		}
		if subscribed {
			log4plus.Info("%s subscribed is true", funName)
			return nil, 100
		} else {
			log4plus.Info("%s subscribed is false", funName)
			return nil, 0
		}
	}
	return nil, 0
}

func (p *UserDB) CheckDiscordId(discordId string) (error, int) {
	funName := "CheckDiscordId"
	if !p.mysqlDb.IsConnect() {
		errString := fmt.Sprintf("%s Db Not Connect ---->>>>", funName)
		log4plus.Error(errString)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName)), 0
	}
	sql := fmt.Sprintf(`select IFNULL(id,-1), subscribed, IFNULL(createtime,''), IFNULL(lasttime,''), IFNULL(state,0) from user where discordid='%s';`, discordId)
	log4plus.Info("%s Query SQL=[%s]", funName, sql)
	rows, err := p.mysqlDb.Mysqldb.Query(sql)
	if err != nil {
		errString := fmt.Sprintf("%s Query Failed Error=[%s] SQL=[%s]", funName, err.Error(), sql)
		log4plus.Error(errString)
		return err, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id, state int64
		var subscribed bool
		var createtime, lasttime string
		scanErr := rows.Scan(&id, &subscribed, &createtime, &lasttime, &state)
		if scanErr != nil {
			log4plus.Error("%s Scan Error=[%s]", funName, scanErr.Error())
			continue
		}
		if subscribed {
			return nil, 100
		} else {
			return nil, 0
		}
	}
	return nil, 100
}

func (p *UserDB) GetApiKey(discordId string) (error, string) {
	funName := "GetApiKey"
	if !p.mysqlDb.IsConnect() {
		errString := fmt.Sprintf("%s Db Not Connect ---->>>>", funName)
		log4plus.Error(errString)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName)), ""
	}
	sql := fmt.Sprintf(`select 	IFNULL(id,-1), IFNULL(apiKey,'') from user where discordid='%s' limit 1;`, discordId)
	log4plus.Info("%s Query SQL=[%s]", funName, sql)
	rows, err := p.mysqlDb.Mysqldb.Query(sql)
	if err != nil {
		errString := fmt.Sprintf("%s Query Failed Error=[%s] SQL=[%s]", funName, err.Error(), sql)
		log4plus.Error(errString)
		return err, ""
	}
	defer rows.Close()

	var apiKey string = ""
	for rows.Next() {
		var tmpId int64
		scanErr := rows.Scan(&tmpId, &apiKey)
		if scanErr != nil {
			log4plus.Error("%s Scan Error=[%s]", funName, scanErr.Error())
			continue
		}
		return nil, apiKey
	}
	errString := fmt.Sprintf("%s Query Failed SQL=[%s]", funName, sql)
	return errors.New(errString), ""
}

func SingtonUserDB() *UserDB {
	if gUserDB == nil {
		log4plus.Info("SingtonUserDB ---->>>>")
		gUserDB = &UserDB{}
		if gUserDB.mysqlDb = NewMysql(configure.SingtonConfigure().Mysql.MysqlIp,
			configure.SingtonConfigure().Mysql.MysqlPort,
			configure.SingtonConfigure().Mysql.MysqlDBName,
			configure.SingtonConfigure().Mysql.MysqlDBCharset,
			configure.SingtonConfigure().Mysql.UserName,
			configure.SingtonConfigure().Mysql.Password); gUserDB.mysqlDb == nil {
			log4plus.Error("SingtonUserDB NewMysql Failed")
			return nil
		}
	}
	return gUserDB
}
