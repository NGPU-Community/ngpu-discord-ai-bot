package db

import (
	"errors"
	"fmt"
	"time"

	log4plus "github.com/nGPU/common/log4go"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type MysqlManager struct {
	connected       bool
	MysqlIp         string
	MysqlPort       int
	MysqlDBName     string
	MysqlDBCharset  string
	UserName        string
	Password        string
	disconnectTimer int64
	Mysqldb         *sqlx.DB
}

// 连接Mysql
func (m *MysqlManager) connectMysql() bool {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&multiStatements=true", m.UserName,
		m.Password,
		m.MysqlIp,
		m.MysqlPort,
		m.MysqlDBName,
		m.MysqlDBCharset)
	if m.Mysqldb == nil {
		log4plus.Info("connectMysql Mysqldb is nil dsn=%s", dsn)
		var err error
		if m.Mysqldb, err = sqlx.Connect("mysql", dsn); err != nil {
			log4plus.Error("mySQL Connect Failed dsn=%s", dsn)
			m.connected = false
			return false
		}
		m.Mysqldb.SetMaxOpenConns(20)
		m.Mysqldb.SetMaxIdleConns(10)
		m.Mysqldb.SetConnMaxLifetime(time.Hour)
		m.connected = true
		log4plus.Info("connectMysql connected is ture")
	} else {
		log4plus.Info("connectMysql Mysqldb is nil dsn=%s", dsn)
		if m.connected {
			if err := m.Mysqldb.Ping(); err != nil {
				m.connected = false
				return false
			}
		} else {
			log4plus.Info("connectMysql Mysqldb Close->Open")
			var err error
			m.Mysqldb.Close()
			if m.Mysqldb, err = sqlx.Open("mysql", dsn); err != nil {
				log4plus.Error("mySQL Connect Failed dsn=%s", dsn)
				m.connected = false
				return false
			}
			if err = m.Mysqldb.Ping(); err != nil {
				m.connected = false
				return false
			}
			m.connected = true
		}
	}
	return m.connected
}

// 连接Mysql
func (m *MysqlManager) IsConnect() bool {
	return m.connected
}

// 检测Mysql连接
func (m *MysqlManager) checkMysql() error {
	if m.Mysqldb == nil {
		return errors.New("m.Mysqldb is Nil")
	}
	_, err := m.Mysqldb.Query("select * from `subscription` limit 1;")
	if err != nil {
		log4plus.Error("Check Mysql Failed, err=%s", err.Error())
		return err
	}
	return nil
}

func (m *MysqlManager) pollMysql() {
	//每5分钟，查询一次Mysql是否可以使用;
	for {
		//判断链接数据库是否正常
		if m.Mysqldb != nil {
			rows, err := m.Mysqldb.Query("select * from `subscription` limit 1;")
			if err != nil {
				m.connected = false
				m.disconnectTimer = time.Now().Unix()
				log4plus.Error("Check Mysql Connect Failed err=%s", err.Error())
			} else {
				rows.Close()
			}
		}

		//判断重新链接
		if !m.IsConnect() {
			if time.Now().Unix()-m.disconnectTimer >= 300 {
				//重新链接
				m.connected = m.connectMysql()
				if !m.connected {
					m.disconnectTimer = time.Now().Unix()
					log4plus.Error("ReConnect Mysql Failed")
				}
			}
		} else {
			// log4plus.Info("Check Mysql Connected Success")
		}
		time.Sleep(30 * time.Second)
	}

}

func NewMysql(Ip string, Port int, DBName string, DBCharset string, UserName string, Password string) *MysqlManager {
	sql := &MysqlManager{
		MysqlIp:        Ip,
		MysqlPort:      Port,
		MysqlDBName:    DBName,
		MysqlDBCharset: DBCharset,
		UserName:       UserName,
		Password:       Password,
		connected:      false,
	}
	go sql.pollMysql()
	return sql
}
