package models

import (
	"fmt"
	"strconv"
	"time"

	"github.com/phachon/mm-wiki/app/utils"
	"github.com/snail007/go-activerecord/mysql"
)

const (
	LogDocument_Action_Create  = 1
	LogDocument_Action_Update  = 2
	LogDocument_Action_Delete  = 3
	LogDocument_Action_Read    = 4
	LogDocument_Action_Recover = 5 // 回收站恢复文档
	LogDocument_Action_PermDel = 6 // 彻底删除文档
	LogDocument_Action_Clear   = 7 // 清空回收站
)

const Table_LogDocument_Name = "log_document"

type LogDocument struct {
}

var LogDocumentModel = LogDocument{}

func (ld *LogDocument) GetLogDocumentByLogDocumentId(logDocId string) (logDocuments map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_LogDocument_Name).Where(map[string]interface{}{
		"log_document_id": logDocId,
	}))
	if err != nil {
		return
	}
	logDocuments = rs.Row()
	return
}

func (ld *LogDocument) Insert(logDocument map[string]interface{}) (id int64, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Exec(db.AR().Insert(Table_LogDocument_Name, logDocument))
	if err != nil {
		return
	}
	id = rs.LastInsertId
	return
}

func (ld *LogDocument) CreateAction(userId string, documentId string, spaceId string) (id int64, err error) {
	logDocument := map[string]interface{}{
		"user_id":     userId,
		"document_id": documentId,
		"space_id":    spaceId,
		"comment":     "创建了文档",
		"action":      LogDocument_Action_Create,
		"create_time": time.Now().Unix(),
	}
	return ld.Insert(logDocument)
}

func (ld *LogDocument) UpdateAction(userId string, documentId string, comment string, spaceId string) (id int64, err error) {
	logDocument := map[string]interface{}{
		"user_id":     userId,
		"document_id": documentId,
		"space_id":    spaceId,
		"comment":     comment,
		"action":      LogDocument_Action_Update,
		"create_time": time.Now().Unix(),
	}
	return ld.Insert(logDocument)
}

func (ld *LogDocument) DeleteAction(userId string, documentId string, spaceId string) (id int64, err error) {
	logDocument := map[string]interface{}{
		"user_id":     userId,
		"document_id": documentId,
		"space_id":    spaceId,
		"comment":     "删除了该文档",
		"action":      LogDocument_Action_Delete,
		"create_time": time.Now().Unix(),
	}
	return ld.Insert(logDocument)
}

func (ld *LogDocument) ReadAction(userId string, documentId string, spaceId string) (id int64, err error) {

	//十分钟内重复阅读不做入库
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().Select("count(*) as count").From(Table_LogDocument_Name).Where(map[string]interface{}{
		"document_id":    documentId,
		"user_id":        userId,
		"action":         LogDocument_Action_Read,
		"create_time >=": time.Now().Unix() - 600,
	}))
	if err != nil {
		return
	}
	count := utils.Convert.StringToInt64(rs.Value("count"))

	if count < 1 {
		logDocument := map[string]interface{}{
			"user_id":     userId,
			"document_id": documentId,
			"space_id":    spaceId,
			"comment":     "阅读了文档",
			"action":      LogDocument_Action_Read,
			"create_time": time.Now().Unix(),
		}
		return ld.Insert(logDocument)
	}
	return
}

func (ld *LogDocument) RecoverAction(userId string, documentId string, spaceId string) (id int64, err error) {
	logDocument := map[string]interface{}{
		"user_id":     userId,
		"document_id": documentId,
		"space_id":    spaceId,
		"comment":     "从回收站恢复了该文档",
		"action":      LogDocument_Action_Recover,
		"create_time": time.Now().Unix(),
	}
	return ld.Insert(logDocument)
}

func (ld *LogDocument) PermanentlyDeleteAction(userId string, documentId string, spaceId string) (id int64, err error) {
	logDocument := map[string]interface{}{
		"user_id":     userId,
		"document_id": documentId,
		"space_id":    spaceId,
		"comment":     "彻底删除了该文档（回收站）",
		"action":      LogDocument_Action_PermDel,
		"create_time": time.Now().Unix(),
	}
	return ld.Insert(logDocument)
}

func (ld *LogDocument) ClearRecycleAction(userId string, spaceId string, count int) (id int64, err error) {
	logDocument := map[string]interface{}{
		"user_id":     userId,
		"document_id": "0",
		"space_id":    spaceId,
		"comment":     fmt.Sprintf("清空了回收站，共 %d 个文档", count),
		"action":      LogDocument_Action_Clear,
		"create_time": time.Now().Unix(),
	}
	return ld.Insert(logDocument)
}

func (ld *LogDocument) GetLogDocumentsByDocumentId(documentId string) (logDocuments []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_LogDocument_Name).Where(map[string]interface{}{
		"document_id": documentId,
	}))
	if err != nil {
		return
	}
	logDocuments = rs.Rows()
	return
}

func (ld *LogDocument) GetLogDocumentsByUserId(userId string) (logDocuments []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_LogDocument_Name).Where(map[string]interface{}{
		"user_id": userId,
	}))
	if err != nil {
		return
	}
	logDocuments = rs.Rows()
	return
}

func (ld *LogDocument) GetLogDocumentsByDocumentIdAndLimit(documentId string, limit int, number int, logReadActionStatus bool) (logDocuments []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet

	where := map[string]interface{}{
		"document_id": documentId,
	}
	if logReadActionStatus {
		where["action"] = LogDocument_Action_Read
	} else {
		where["action !="] = LogDocument_Action_Read
	}

	rs, err = db.Query(db.AR().From(Table_LogDocument_Name).Where(where).Limit(limit, number).OrderBy("log_document_id", "DESC"))
	if err != nil {
		return
	}
	logDocuments = rs.Rows()

	return
}

func (ld *LogDocument) GetLogDocumentLastUpdateByDocumentId(documentId string) (logDocuments map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet

	where := map[string]interface{}{
		"document_id": documentId,
	}
	where["action"] = LogDocument_Action_Update

	rs, err = db.Query(db.AR().From(Table_LogDocument_Name).Where(where).OrderBy("log_document_id", "DESC"))
	if err != nil {
		return
	}
	logDocuments = rs.Row()

	return
}

func (ld *LogDocument) GetLogDocumentsByUserIdAndLimit(userId string, limit int, number int) (logDocuments []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_LogDocument_Name).Where(map[string]interface{}{
		"user_id": userId,
	}).Limit(limit, number).OrderBy("log_document_id", "DESC"))
	if err != nil {
		return
	}
	logDocuments = rs.Rows()

	return
}

func (ld *LogDocument) GetLogDocumentsByUserIdKeywordAndLimit(userId string, keyword string, limit int, number int) (logDocuments []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_LogDocument_Name).Where(map[string]interface{}{
		"comment LIKE": "%" + keyword + "%",
		"user_id":      userId,
	}).Limit(limit, number).OrderBy("log_document_id", "DESC"))
	if err != nil {
		return
	}
	logDocuments = rs.Rows()

	return
}

func (ld *LogDocument) GetLogDocumentsByKeywordAndLimit(keyword string, limit int, number int) (logDocuments []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_LogDocument_Name).Where(map[string]interface{}{
		"comment LIKE": "%" + keyword + "%",
	}).Limit(limit, number).OrderBy("log_document_id", "DESC"))
	if err != nil {
		return
	}
	logDocuments = rs.Rows()

	return
}

func (ld *LogDocument) GetDefaultLogDocumentsByLimit(userId string, limit int, number int, limitUser bool) (logDocuments []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	where := db.AR().From(Table_LogDocument_Name)

	var speaceIds []string

	// 查询用户空间权限
	spaceUserRs, err := db.Query(db.AR().From(Table_SpaceUser_Name).Where(map[string]interface{}{
		"user_id": userId,
	}))

	spaceUsers := spaceUserRs.Rows()
	spaceUsersLen := len(spaceUsers)
	for i := 0; i < spaceUsersLen; i++ {
		spaceUser := spaceUsers[i]
		speaceIds = append(speaceIds, spaceUser["space_id"])
	}

	// 查询公共空间
	spaceRs, err := db.Query(db.AR().From(Table_Space_Name).Where(map[string]interface{}{
		"visit_level": "public",
	}))

	spaces := spaceRs.Rows()
	for i := 0; i < len(spaces); i++ {
		space := spaces[i]
		speaceIds = append(speaceIds, space["space_id"])
	}

	where.WhereWrap(map[string]interface{}{
		"action != ": LogDocument_Action_Read,
	}, "", " and ")

	if limitUser == true {
		where.WhereWrap(map[string]interface{}{
			"user_id": userId,
		}, "", " and ")
	}

	where.WhereWrap(map[string]interface{}{
		"space_id": speaceIds,
	}, "", "")

	rs, err = db.Query(where.Limit(limit, number).OrderBy("log_document_id", "DESC"))

	if err != nil {
		return
	}
	logDocuments = rs.Rows()

	return
}

func (ld *LogDocument) GetLogDocumentsByLimit(userId string, limit int, number int) (logDocuments []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	where := db.AR().From(Table_LogDocument_Name)

	// 查询用户空间权限
	spaceUserRs, err := db.Query(db.AR().From(Table_SpaceUser_Name).Where(map[string]interface{}{
		"user_id": userId,
	}))

	spaceUsers := spaceUserRs.Rows()
	spaceUsersLen := len(spaceUsers)

	for i := 0; i < spaceUsersLen; i++ {
		spaceUser := spaceUsers[i]
		if i == 0 {
			where.WhereWrap(map[string]interface{}{
				"space_id": spaceUser["space_id"],
			}, "", "")
		} else {
			where.WhereWrap(map[string]interface{}{
				"space_id": spaceUser["space_id"],
			}, "or", "")
		}
	}

	// 查询公共空间
	spaceRs, err := db.Query(db.AR().From(Table_Space_Name).Where(map[string]interface{}{
		"visit_level": "public",
	}))

	spaces := spaceRs.Rows()

	for i := 0; i < len(spaces); i++ {
		space := spaces[i]
		if i == 0 && spaceUsersLen == 0 {
			where.WhereWrap(map[string]interface{}{
				"space_id": space["space_id"],
			}, "", "")
		} else {
			where.WhereWrap(map[string]interface{}{
				"space_id": space["space_id"],
			}, "or", "")
		}
	}

	rs, err = db.Query(where.Limit(limit, number).OrderBy("log_document_id", "DESC"))
	if err != nil {
		return
	}
	logDocuments = rs.Rows()

	return
}

func (ld *LogDocument) CountLogDocumentsByDocumentId(documentId string, logReadActionStatus bool) (count int64, err error) {

	db := G.DB()
	var rs *mysql.ResultSet

	where := map[string]interface{}{
		"document_id": documentId,
	}
	if logReadActionStatus {
		where["action"] = LogDocument_Action_Read
	} else {
		where["action !="] = LogDocument_Action_Read
	}

	rs, err = db.Query(db.AR().Select("count(*) as total").From(Table_LogDocument_Name).Where(where))
	if err != nil {
		return
	}
	count = utils.Convert.StringToInt64(rs.Value("total"))
	return
}

func (ld *LogDocument) CountReadActionLogDocumentsByDocumentId(documentId string) (logReadActionCounts map[string]interface{}, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().Select("count(*) as total, COUNT(DISTINCT user_id) as num").From(Table_LogDocument_Name).Where(map[string]interface{}{
		"document_id": documentId,
		"action":      LogDocument_Action_Read,
	}))
	if err != nil {
		return
	}
	count := utils.Convert.StringToInt64(rs.Value("total"))
	num := utils.Convert.StringToInt64(rs.Value("num"))

	var rs7day *mysql.ResultSet
	rs7day, err = db.Query(db.AR().Select("count(*) as total, COUNT(DISTINCT user_id) as num").From(Table_LogDocument_Name).Where(map[string]interface{}{
		"document_id":    documentId,
		"action":         LogDocument_Action_Read,
		"create_time >=": time.Now().Unix() - (7 * 86400),
	}))
	if err != nil {
		return
	}

	count7day := utils.Convert.StringToInt64(rs7day.Value("total"))
	num7day := utils.Convert.StringToInt64(rs7day.Value("num"))

	logReadActionCounts = map[string]interface{}{
		"count":     count,
		"num":       num,
		"count7day": count7day,
		"num7day":   num7day,
	}
	return
}

func (ld *LogDocument) CountLogDocumentsByUserId(userId string) (count int64, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().Select("count(*) as total").From(Table_LogDocument_Name).Where(map[string]interface{}{
		"user_id": userId,
	}))
	if err != nil {
		return
	}
	count = utils.Convert.StringToInt64(rs.Value("total"))
	return
}

func (ld *LogDocument) CountLogDocumentsByUserIdAndKeyword(userId string, keyword string) (count int64, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().Select("count(*) as total").From(Table_LogDocument_Name).Where(map[string]interface{}{
		"comment LIKE": "%" + keyword + "%",
		"user_id":      userId,
	}))
	if err != nil {
		return
	}
	count = utils.Convert.StringToInt64(rs.Value("total"))
	return
}

func (ld *LogDocument) CountLogDocumentsByKeyword(keyword string) (count int64, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().Select("count(*) as total").From(Table_LogDocument_Name).Where(map[string]interface{}{
		"comment LIKE": "%" + keyword + "%",
	}))
	if err != nil {
		return
	}
	count = utils.Convert.StringToInt64(rs.Value("total"))
	return
}

func (ld *LogDocument) CountLogDocuments() (count int64, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(
		db.AR().
			Select("count(*) as total").
			From(Table_LogDocument_Name))
	if err != nil {
		return
	}
	count = utils.Convert.StringToInt64(rs.Value("total"))
	return
}

// AttachDocuments 给日志列表关联文档名、文档类型、用户名等信息
// 文档已被彻底删除时，根据 action 类型设置合理的默认值
func (ld *LogDocument) AttachDocuments(logDocuments []map[string]string, users []map[string]string, docs []map[string]string) {

	// 构建 user map
	userMap := make(map[string]map[string]string)
	for _, user := range users {
		userMap[user["user_id"]] = user
	}
	// 构建 doc map
	docMap := make(map[string]map[string]string)
	for _, doc := range docs {
		docMap[doc["document_id"]] = doc
	}
	// 构建 space map（用于清空回收站显示空间名称）
	spaceIds := []string{}
	for _, logDocument := range logDocuments {
		if logDocument["space_id"] != "" {
			spaceIds = append(spaceIds, logDocument["space_id"])
		}
	}
	spaceMap := make(map[string]map[string]string)
	if len(spaceIds) > 0 {
		spaces, _ := SpaceModel.GetSpaceBySpaceIds(spaceIds)
		for _, space := range spaces {
			spaceMap[space["space_id"]] = space
		}
	}

	for _, logDocument := range logDocuments {
		// 用户信息
		logDocument["username"] = ""
		logDocument["given_name"] = ""
		if user, ok := userMap[logDocument["user_id"]]; ok {
			logDocument["username"] = user["username"]
			logDocument["given_name"] = user["given_name"]
		}
		// 根据 action 设置默认文档名（action 从数据库查出的是字符串，转成 int 跟常量比较）
		actionInt, _ := strconv.Atoi(logDocument["action"])
		if actionInt == LogDocument_Action_Clear { // 清空回收站，空间级操作，显示空间名称
			if space, ok := spaceMap[logDocument["space_id"]]; ok {
				logDocument["document_name"] = "清空回收站(" + space["name"] + ")"
			} else {
				logDocument["document_name"] = "清空回收站(space_id:" + logDocument["space_id"] + ")"
			}
		} else if actionInt == LogDocument_Action_PermDel { // 彻底删除
			logDocument["document_name"] = "已彻底删除文档(" + logDocument["document_id"] + ")"
		} else { // 普通删除
			logDocument["document_name"] = "已删除文档(" + logDocument["document_id"] + ")"
		}
		logDocument["document_type"] = "1"
		// 如果文档存在，用真实信息覆盖
		if doc, ok := docMap[logDocument["document_id"]]; ok {
			logDocument["document_name"] = doc["name"]
			logDocument["document_type"] = doc["type"]
			logDocument["update_time"] = doc["update_time"]
		}
	}
}
