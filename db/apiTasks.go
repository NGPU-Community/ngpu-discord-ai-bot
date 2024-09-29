package db

import (
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"github.com/nGPU/bot/header"

	"github.com/nGPU/bot/configure"
	log4plus "github.com/nGPU/common/log4go"
)

type APITasksDB struct {
	outChan chan bool
	mysqlDb *MysqlManager
	roots   *x509.CertPool
	rootPEM []byte
}

var gAPITasksDB *APITasksDB

// 插入Ai任务
func (p *APITasksDB) InsertAiTask(taskId, apiKey, request, requestTime, gpuUrl, method string) error {
	funName := "InsertAiTask"
	if !p.mysqlDb.IsConnect() {
		log4plus.Error("%s Db Not Connect ---->>>>", funName)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName))
	}
	sql := fmt.Sprintf(`select IFNULL(id,-1) from apitasks where taskId='%s' limit 1;`, taskId)
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
		sql := fmt.Sprintf(`insert into apitasks (taskId, apiKey, request, requestTime, state, gpuDuration, gpuUrl, method) values ('%s', '%s', ?, '%s', 0, 0, '%s', '%s');`,
			taskId,
			apiKey,
			requestTime,
			gpuUrl,
			method)
		log4plus.Info("%s Insert SQL=[%s]", funName, sql)
		_, err := p.mysqlDb.Mysqldb.Exec(sql, request)
		if err != nil {
			log4plus.Error("%s Exec err=%s taskId=%s", funName, err.Error(), taskId)
			return err
		}
		return nil
	} else {
		errString := fmt.Sprintf("%s Failed to create apitasks, taskId already exists taskId=[%s]", funName, taskId)
		log4plus.Error(errString)
		return errors.New(errString)
	}
}

func (p *APITasksDB) SetAiTaskRunning(taskId, response, responseTime string, gpuDuration int) error {
	funName := "SetAiTask"
	if !p.mysqlDb.IsConnect() {
		errString := fmt.Sprintf("%s Db Not Connect ---->>>>", funName)
		log4plus.Error(errString)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName))
	}

	sql := fmt.Sprintf(`Update apitasks Set response=?, responseTime='%s', state=%d, gpuDuration=%d where taskid='%s';`,
		responseTime, header.RunningState, gpuDuration, taskId)

	log4plus.Info("%s SQL=[%s]", funName, sql)
	_, err := p.mysqlDb.Mysqldb.Exec(sql, response)
	if err != nil {
		log4plus.Error("%s Exec err=%s taskId=%s", funName, err.Error(), taskId)
		return err
	}
	return nil
}

func (p *APITasksDB) SetAiTaskSuccess(taskId, response, responseTime string, gpuDuration int) error {
	funName := "SetAiTask"
	if !p.mysqlDb.IsConnect() {
		errString := fmt.Sprintf("%s Db Not Connect ---->>>>", funName)
		log4plus.Error(errString)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName))
	}

	sql := fmt.Sprintf(`Update apitasks Set response=?, responseTime='%s', state=%d, gpuDuration=%d where taskid='%s';`,
		responseTime, header.FinishState, gpuDuration, taskId)

	log4plus.Info("%s SQL=[%s]", funName, sql)
	_, err := p.mysqlDb.Mysqldb.Exec(sql, response)
	if err != nil {
		log4plus.Error("%s Exec err=%s taskId=%s", funName, err.Error(), taskId)
		return err
	}
	return nil
}

func (p *APITasksDB) SetAiTaskFail(taskId, response, responseTime string) error {
	funName := "SetAiTask"
	if !p.mysqlDb.IsConnect() {
		errString := fmt.Sprintf("%s Db Not Connect ---->>>>", funName)
		log4plus.Error(errString)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName))
	}

	sql := fmt.Sprintf(`Update apitasks Set response=?, responseTime='%s', state=%d where taskid='%s';`,
		responseTime, header.ErrorState, taskId)

	log4plus.Info("%s SQL=[%s]", funName, sql)
	_, err := p.mysqlDb.Mysqldb.Exec(sql, response)
	if err != nil {
		log4plus.Error("%s Exec err=%s taskId=%s", funName, err.Error(), taskId)
		return err
	}
	return nil
}

func (p *APITasksDB) GetAiTasks(methods []string) (error, []*header.FaceFusionDBData) {
	funName := "GetAiTasks"
	if !p.mysqlDb.IsConnect() {
		errString := fmt.Sprintf("%s Db Not Connect ---->>>>", funName)
		log4plus.Error(errString)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName)), nil
	}
	methodSql := ""
	for index, v := range methods {
		if index == 0 {
			methodSql = methodSql + fmt.Sprintf(" (method='%s'", v)
			if len(methods) == 1 {
				methodSql = methodSql + ")"
			}
		} else if index == len(methods)-1 {
			methodSql = methodSql + fmt.Sprintf(" or method='%s')", v)
		} else {
			methodSql = methodSql + fmt.Sprintf(" or method='%s'", v)
		}
	}
	sql := fmt.Sprintf(`select IFNULL(id,-1), IFNULL(taskId,''), IFNULL(requesttime,''), IFNULL(method, ''), IFNULL(response,'') from apitasks where state=%d and %s order by id limit 100;`, header.RunningState, methodSql)
	// log4plus.Info("%s Query SQL=[%s]", funName, sql)
	rows, err := p.mysqlDb.Mysqldb.Query(sql)
	if err != nil {
		log4plus.Error("%s Query Failed Error=[%s] SQL=[%s]", funName, err.Error(), sql)
		return err, nil
	}
	defer rows.Close()

	var faceFusionTasks []*header.FaceFusionDBData
	for rows.Next() {
		var id int
		var taskId, method, response string
		var requestTime string
		scanErr := rows.Scan(&id, &taskId, &requestTime, &method, &response)
		if scanErr != nil {
			log4plus.Error("%s Scan Error=%s", funName, scanErr.Error())
			continue
		}
		log4plus.Info("%s rows.Scan id=[%d]", funName, id)
		if id == -1 {
			continue
		}
		faceFusionTasks = append(faceFusionTasks, &header.FaceFusionDBData{
			TaskID:       taskId,
			RequestTime:  requestTime,
			Method:       method,
			ResponseData: response,
		})
	}
	return nil, faceFusionTasks
}

func (p *APITasksDB) GetTaskId(taskId string) (error, *header.TaskInfo) {
	funName := "GetTaskId"
	if !p.mysqlDb.IsConnect() {
		errString := fmt.Sprintf("%s Db Not Connect ---->>>>", funName)
		log4plus.Error(errString)
		return errors.New(fmt.Sprintf("%s Db Not Connect", funName)), nil
	}
	sql := fmt.Sprintf(`select IFNULL(id,-1), IFNULL(taskId,''), IFNULL(requesttime,''), IFNULL(method, ''), IFNULL(response,''), IFNULL(state, -1), IFNULL(responseTime,'') from apitasks where taskId='%s';`, taskId)
	log4plus.Info("%s Query SQL=[%s]", funName, sql)
	rows, err := p.mysqlDb.Mysqldb.Query(sql)
	if err != nil {
		log4plus.Error("%s Query Failed Error=[%s] SQL=[%s]", funName, err.Error(), sql)
		return err, nil
	}
	defer rows.Close()

	for rows.Next() {
		var id, state int
		var requestTime, method, response, reposneTime string
		scanErr := rows.Scan(&id, &taskId, &requestTime, &method, &response, &state, &reposneTime)
		if scanErr != nil {
			log4plus.Error("%s Scan Error=%s", funName, scanErr.Error())
			continue
		}
		log4plus.Info("%s rows.Scan id=[%d]", funName, id)
		if id == -1 {
			continue
		}
		// 将字符串转换为 time.Time 类型
		reqTime, err := time.Parse("2006-01-02 15:04:05", requestTime)
		if err != nil {
			log4plus.Error("%s time.Parse requestTime err=[%s]", funName, err.Error())
			continue
		}
		var resTime time.Time
		if state == header.FinishState {
			resTime, err = time.Parse("2006-01-02 15:04:05", reposneTime)
			if err != nil {
				log4plus.Error("%s time.Parse reposneTime err=[%s]", funName, err.Error())
				continue
			}
		} else {
			resTime = time.Now()
		}
		duration := resTime.Sub(reqTime)
		return nil, &header.TaskInfo{
			TaskId:        taskId,
			RequestTime:   requestTime,
			Method:        method,
			Response:      response,
			State:         state,
			RecordDursion: int64(duration.Seconds()),
		}
	}
	log4plus.Error("%s not found taskId=[%s]", funName, taskId)
	return errors.New("not found taskId"), nil
}

func SingtonAPITasksDB() *APITasksDB {
	if gAPITasksDB == nil {
		log4plus.Info("SingtonUAPITasksDB ---->>>>")
		gAPITasksDB = &APITasksDB{}
		if gAPITasksDB.mysqlDb = NewMysql(configure.SingtonConfigure().Mysql.MysqlIp,
			configure.SingtonConfigure().Mysql.MysqlPort,
			configure.SingtonConfigure().Mysql.MysqlDBName,
			configure.SingtonConfigure().Mysql.MysqlDBCharset,
			configure.SingtonConfigure().Mysql.UserName,
			configure.SingtonConfigure().Mysql.Password); gAPITasksDB.mysqlDb == nil {
			log4plus.Error("SingtonUAPITasksDB NewMysql Failed")
			return nil
		}
	}
	return gAPITasksDB
}
