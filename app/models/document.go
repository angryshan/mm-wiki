package models

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/phachon/mm-wiki/app/utils"
	"github.com/snail007/go-activerecord/mysql"
	"strconv"
	"strings"
	"time"
)

const (
	Document_Delete_True  = 1
	Document_Delete_False = 0

	Document_Type_Page = 1
	Document_Type_Dir  = 2
)

const Table_Document_Name = "document"

type Document struct {
}

var DocumentModel = Document{}

// get document by document_id
func (d *Document) GetDocumentByDocumentId(documentId string) (document map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"document_id": documentId,
		"is_delete":   Document_Delete_False,
	}))
	if err != nil {
		return
	}
	document = rs.Row()
	return
}

// get documents by parent_id
func (d *Document) GetDocumentsByParentId(parentId string) (documents []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"parent_id": parentId,
		"is_delete": Document_Delete_False,
	}))
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

// get document by name
func (d *Document) GetDocumentsByName(name string) (documents []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"name":      name,
		"is_delete": Document_Delete_False,
	}))
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

// get document by name and spaceId
func (d *Document) GetDocumentByNameAndSpaceId(name string, spaceId string) (document map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"name":      name,
		"space_id":  spaceId,
		"is_delete": Document_Delete_False,
	}).Limit(0, 1))
	if err != nil {
		return
	}
	document = rs.Row()
	return
}

// get document by name and spaceId
func (d *Document) GetDocumentByNameParentIdAndSpaceId(name string, parentId string, spaceId string, docType int) (document map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"name":      name,
		"space_id":  spaceId,
		"parent_id": parentId,
		"type":      docType,
		"is_delete": Document_Delete_False,
	}).Limit(0, 1))
	if err != nil {
		return
	}
	document = rs.Row()
	return
}

// get document by name and spaceId
func (d *Document) GetDocumentByParentIdAndSpaceId(parentId string, spaceId string, docType int) (document map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"space_id":  spaceId,
		"parent_id": parentId,
		"type":      docType,
		"is_delete": Document_Delete_False,
	}).Limit(0, 1))
	if err != nil {
		return
	}
	document = rs.Row()
	return
}

// get document by name and spaceId
func (d *Document) GetDocumentsByParentIdAndSpaceIdOnly(parentId string, spaceId string) (document []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"space_id":  spaceId,
		"parent_id": parentId,
		"is_delete": Document_Delete_False,
	}))
	if err != nil {
		return
	}
	document = rs.Rows()
	return
}

// get max sequence
func (d *Document) GetDocumentMaxSequence(parentId string, spaceId string) (sequence int, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"space_id":  spaceId,
		"parent_id": parentId,
	}).OrderBy("sequence", "desc").Limit(0, 1))

	if err != nil {
		return
	}

	document := rs.Row()
	sequenceStr := document["sequence"]
	sequence, err = strconv.Atoi(sequenceStr)
	return
}

// delete document by document_id
func (d *Document) DeleteDBAndFile(documentId string, spaceId string, userId string, pageFile string, docType string) (err error) {
	db := G.DB()
	tx, err := db.Begin(db.Config)
	if err != nil {
		return
	}
	_, err = db.ExecTx(db.AR().Update(Table_Document_Name, map[string]interface{}{
		"is_delete":    Document_Delete_True,
		"update_time":  time.Now().Unix(),
		"edit_user_id": userId,
	}, map[string]interface{}{
		"document_id": documentId,
	}), tx)
	if err != nil {
		tx.Rollback()
		return
	}

	// delete document file
	err = utils.Document.Delete(pageFile, utils.Convert.StringToInt(docType))
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	// create document log
	go func(userId, documentId string, spaceId string) {
		_, err := LogDocumentModel.DeleteAction(userId, documentId, spaceId)
		if err != nil {
			logs.Error("delete document add log err=%s", err.Error())
		}
	}(userId, documentId, spaceId)

	// delete follow doc
	go func(documentId string) {
		err := FollowModel.DeleteByObjectIdType(documentId, fmt.Sprintf("%d", Follow_Type_Doc))
		if err != nil {
			logs.Error("delete follow document err=%s", err.Error())
		}
	}(documentId)

	// delete collect doc
	go func(documentId string) {
		err := CollectionModel.DeleteByResourceIdType(documentId, fmt.Sprintf("%d", Collection_Type_Doc))
		if err != nil {
			logs.Error("delete collect document err=%s", err.Error())
		}
	}(documentId)

	return
}

// insert document
func (d *Document) Insert(documentValue map[string]interface{}) (id int64, err error) {

	db := G.DB()
	// start db begin
	tx, err := db.Begin(db.Config)
	if err != nil {
		return
	}

	// 处理同级排序编号
	parentId := documentValue["parent_id"].(string)
	spaceId := documentValue["space_id"].(string)

	sequence, err := d.GetDocumentMaxSequence(parentId, spaceId)
	if err != nil {
		sequence = 0
	}

	sequence += 1
	documentValue["sequence"] = strconv.Itoa(sequence)

	var rs *mysql.ResultSet
	documentValue["create_time"] = time.Now().Unix()
	documentValue["update_time"] = time.Now().Unix()
	rs, err = db.ExecTx(db.AR().Insert(Table_Document_Name, documentValue), tx)
	if err != nil {
		tx.Rollback()
		return
	}
	id = rs.LastInsertId

	// create document page file
	document := map[string]string{
		"space_id":  documentValue["space_id"].(string),
		"parent_id": documentValue["parent_id"].(string),
		"name":      documentValue["name"].(string),
		"type":      fmt.Sprintf("%d", documentValue["type"].(int)),
		"path":      documentValue["path"].(string),
	}
	_, pageFile, err := d.GetParentDocumentsByDocument(document)
	err = utils.Document.Create(pageFile)
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	if err != nil {
		return
	}

	userId := documentValue["create_user_id"].(string)
	// create document log
	go func(userId, documentId, spaceId string) {
		_, err := LogDocumentModel.CreateAction(userId, documentId, spaceId)
		if err != nil {
			logs.Error("create document add log err=%s", err.Error())
		}
	}(userId, fmt.Sprintf("%d", id), spaceId)

	// follow document
	go func(userId, documentId string) {
		_, err := FollowModel.CreateAutoFollowDocument(userId, documentId)
		if err != nil {
			logs.Error("follow document err=%s", err.Error())
		}
	}(userId, fmt.Sprintf("%d", id))
	return
}

// update document by document_id
func (d *Document) Update(documentId string, documentValue map[string]interface{}, comment string, spaceId string) (id int64, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	documentValue["update_time"] = time.Now().Unix()
	rs, err = db.Exec(db.AR().Update(Table_Document_Name, documentValue, map[string]interface{}{
		"document_id": documentId,
		"is_delete":   Document_Delete_False,
	}))
	if err != nil {
		return
	}
	id = rs.LastInsertId

	// update document log
	go func(editUserId string, documentId string, comment string, spaceId string) {
		_, err := LogDocumentModel.UpdateAction(editUserId, documentId, comment, spaceId)
		if err != nil {
			logs.Error("update document add log err=%s", err.Error())
		}
	}(documentValue["edit_user_id"].(string), documentId, comment, spaceId)

	// follow document
	go func(editUserId string, documentId string) {
		_, err := FollowModel.CreateAutoFollowDocument(editUserId, documentId)
		if err != nil {
			logs.Error("follow document err=%s", err.Error())
		}
	}(documentValue["edit_user_id"].(string), documentId)
	return
}

// update document by spaceId and >= sequence
// 批量移动更新文档序号
func (d *Document) MoveSequenceBySpaceIdAndGtSequence(spaceId string, startSequence int, n int) (id int64, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	//documentValue := map[string]interface{}{
	//	"sequence": fmt.Sprintf("(sequence+%d)", n),
	//	"update_time": time.Now().Unix(),
	//}
	updateTime := time.Now().Unix()
	sql := fmt.Sprintf(
		"update mw_%s set sequence=sequence+%d, update_time=%d where space_id=%s and sequence >= %d and is_delete=%d",
		Table_Document_Name, n, updateTime, spaceId, startSequence, Document_Delete_False,
	)
	rs, err = db.ExecSQL(sql)
	if err != nil {
		logs.Error(err.Error())
		return
	}
	id = rs.LastInsertId
	return
}

// move document
func (d *Document) MoveDBAndFile(documentId string, spaceId string, updateValue map[string]interface{},
	oldPageFile string, newPageFile string, docType string, comment string) (id int64, err error) {

	db := G.DB()
	tx, err := db.Begin(db.Config)
	if err != nil {
		return
	}
	var rs *mysql.ResultSet
	updateValue["update_time"] = time.Now().Unix()
	rs, err = db.ExecTx(db.AR().Update(Table_Document_Name, updateValue, map[string]interface{}{
		"document_id": documentId,
		"is_delete":   Document_Delete_False,
	}), tx)
	if err != nil {
		tx.Rollback()
		return
	}
	id = rs.LastInsertId

	err = utils.Document.Move(oldPageFile, newPageFile, utils.Convert.StringToInt(docType))
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	if err != nil {
		return
	}

	// create document log
	go func(userId, documentId, comment string, spaceId string) {
		_, err := LogDocumentModel.UpdateAction(updateValue["edit_user_id"].(string), documentId, comment, spaceId)
		if err != nil {
			logs.Error("update document add log err=%s", err.Error())
		}
	}(updateValue["edit_user_id"].(string), documentId, comment, spaceId)

	return
}

// update document by document_id
func (d *Document) UpdateDBAndFile(documentId string, spaceId string, document map[string]string, documentContent string, updateValue map[string]interface{}, comment string) (id int64, err error) {

	// get document page file
	_, oldPageFile, err := DocumentModel.GetParentDocumentsByDocument(document)
	if err != nil {
		return
	}
	// begin update
	db := G.DB()
	tx, err := db.Begin(db.Config)
	if err != nil {
		return
	}
	var rs *mysql.ResultSet
	updateValue["update_time"] = time.Now().Unix()
	rs, err = db.ExecTx(db.AR().Update(Table_Document_Name, updateValue, map[string]interface{}{
		"document_id": documentId,
		"is_delete":   Document_Delete_False,
	}), tx)
	if err != nil {
		tx.Rollback()
		return
	}
	id = rs.LastInsertId

	// update document file
	docType := utils.Convert.StringToInt(document["type"])
	nameIsChange := false
	if updateValue["name"].(string) != document["name"] {
		nameIsChange = true
	}
	err = utils.Document.Update(oldPageFile, updateValue["name"].(string), documentContent, docType, nameIsChange)
	if err != nil {
		tx.Rollback()
		return
	}

	// commit
	err = tx.Commit()
	if err != nil {
		return
	}

	// create document log
	go func(documentId string, comment string, spaceId string) {
		_, _ = LogDocumentModel.UpdateAction(updateValue["edit_user_id"].(string), documentId, comment, spaceId)
	}(documentId, comment, spaceId)

	// create follow doc
	go func(documentId string) {
		_, _ = FollowModel.CreateAutoFollowDocument(updateValue["edit_user_id"].(string), documentId)
	}(documentId)

	return
}

// get all documents
func (d *Document) GetDocumentsBySpaceId(spaceId string) (documents []map[string]string, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(
		db.AR().From(Table_Document_Name).Where(map[string]interface{}{
			"space_id":  spaceId,
			"is_delete": Document_Delete_False,
		}))
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

// get document by spaceId and parentId
func (d *Document) GetDocumentsBySpaceIdAndParentId(spaceId string, parentId string) (documents []map[string]string, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(
		db.AR().From(Table_Document_Name).Where(map[string]interface{}{
			"space_id":  spaceId,
			"parent_id": parentId,
			"is_delete": Document_Delete_False,
		}))
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

// get document by spaceId
func (d *Document) GetSpaceDefaultDocument(spaceId string) (document map[string]string, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(
		db.AR().From(Table_Document_Name).Where(map[string]interface{}{
			"space_id":  spaceId,
			"parent_id": "0",
			"is_delete": Document_Delete_False,
		}).Limit(0, 1))
	if err != nil {
		return
	}
	document = rs.Row()
	return
}

// get document by spaceId
func (d *Document) GetAllSpaceDocuments(spaceId string) (documents []map[string]string, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(
		db.AR().From(Table_Document_Name).Where(map[string]interface{}{
			"space_id":    spaceId,
			"parent_id >": "0",
			"is_delete":   Document_Delete_False,
		}).OrderBy("sequence", "ASC"))
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

// get document count
func (d *Document) CountDocumentsBySpaceId(spaceId string) (count int64, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(
		db.AR().
			Select("count(*) as total").
			From(Table_Document_Name).
			Where(map[string]interface{}{
				"space_id":  spaceId,
				"is_delete": Document_Delete_False,
			}))
	if err != nil {
		return
	}
	count = utils.NewConvert().StringToInt64(rs.Value("total"))
	return
}

// get document count
func (d *Document) CountDocuments() (count int64, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(
		db.AR().
			Select("count(*) as total").
			From(Table_Document_Name).
			Where(map[string]interface{}{
				"is_delete": Document_Delete_False,
			}))
	if err != nil {
		return
	}
	count = utils.NewConvert().StringToInt64(rs.Value("total"))
	return
}

// get document by name
func (d *Document) GetDocumentsByLikeName(name string) (documents []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"name Like": "%" + name + "%",
		"is_delete": Document_Delete_False,
	}))
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

// get document link name and limit
func (d *Document) GetDocumentsByLikeNameAndLimit(name string, limit int, number int) (documents []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"name Like": "%" + name + "%",
		"is_delete": Document_Delete_False,
	}).Limit(limit, number))
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

// count document like name
func (d *Document) CountDocumentsLikeName(name string) (count int64, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(
		db.AR().
			Select("count(*) as total").
			From(Table_Document_Name).
			Where(map[string]interface{}{
				"name Like": "%" + name + "%",
				"is_delete": Document_Delete_False,
			}))
	if err != nil {
		return
	}
	count = utils.Convert.StringToInt64(rs.Value("total"))
	return
}

// get document by spaceId and document_ids
func (d *Document) GetDocumentsByDocumentIds(documentIds []string) (documents []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"document_id": documentIds,
		"is_delete":   Document_Delete_False,
	}))
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

func (d *Document) GetAllDocumentsByDocumentIds(documentIds []string) (documents []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"document_id": documentIds,
	}))
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

func (d *Document) GetAllDocuments() (documents []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"is_delete": Document_Delete_False,
	}))
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

func (d *Document) GetParentDocumentsByDocument(document map[string]string) (parentDocuments []map[string]string, pageFile string, err error) {

	if document["parent_id"] == "0" {
		parentDocuments = append(parentDocuments, document)
		pageFile = utils.Document.GetDefaultPageFileBySpaceName(document["name"])
	} else {
		documentsIds := strings.Split(document["path"], ",")
		parentDocuments, err = d.GetDocumentsByDocumentIds(documentsIds)
		if err != nil {
			return
		}
		var parentPath = ""
		for _, parentDocument := range parentDocuments {
			parentPath += parentDocument["name"] + "/"
		}
		parentPath = strings.TrimRight(parentPath, "/")
		pageFile = utils.Document.GetPageFileByParentPath(document["name"], utils.Convert.StringToInt(document["type"]), parentPath)
	}
	return
}

func (d *Document) GetParentDocumentsByPath(path string) (parentDocuments []map[string]string, err error) {
	documentsIds := strings.Split(path, ",")
	parentDocuments, err = d.GetDocumentsByDocumentIds(documentsIds)
	if err != nil {
		return
	}
	return
}

func (d *Document) GetSpaceIdsOrderByCountDocumentLimit(limit int) (documents []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	sql := db.AR().Select("space_id, count('space_id') as total").
		From(Table_Document_Name).Where(map[string]interface{}{
		"is_delete": Document_Delete_False,
	}).
		GroupBy("space_id").
		OrderBy("total", "DESC").
		Limit(0, limit)
	rs, err = db.Query(sql)
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

func (d *Document) GetCountGroupByCreateTime(startTime int64) (documents []map[string]string, err error) {
	/*select month(FROM_UNIXTIME(time)) from table_name group by month(FROM_UNIXTIME(time))*/

	/*select DATE_FORMAT(FROM_UNIXTIME(time),"%Y-%m") from tcm_fund_list group by DATE_FORMAT(FROM_UNIXTIME(time),"%Y-%m")*/
	/*select FROM_UNIXTIME(time,"%Y-%m") from tcm_fund_list group by FROM_UNIXTIME(time,"%Y-%m")*/

	/*select DATE_FORMAT(FROM_UNIXTIME(time),"%Y-%m-%d") from tcm_fund_list group by DATE_FORMAT(FROM_UNIXTIME(time),"%Y-%m-%d")*/
	db := G.DB()
	var rs *mysql.ResultSet
	sql := db.AR().Select("DATE_FORMAT(FROM_UNIXTIME(create_time),'%Y-%m-%d') as date, count('date') as total").
		From(Table_Document_Name).Where(map[string]interface{}{
		"is_delete":      Document_Delete_False,
		"create_time >=": startTime,
	}).GroupBy("DATE_FORMAT(FROM_UNIXTIME(create_time),'%Y-%m-%d')")
	rs, err = db.Query(sql)
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

func (d *Document) GetDocumentGroupCreateUserId() (documents []map[string]string, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	sql := db.AR().Select("create_user_id, count('create_user_id') as total").
		From(Table_Document_Name).Where(map[string]interface{}{
		"is_delete": Document_Delete_False,
	}).GroupBy("create_user_id")
	rs, err = db.Query(sql)
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

func (d *Document) GetDocumentGroupEditUserId() (documents []map[string]string, err error) {

	db := G.DB()
	var rs *mysql.ResultSet
	sql := db.AR().Select("edit_user_id, count('edit_user_id') as total").
		From(Table_Document_Name).Where(map[string]interface{}{
		"is_delete": Document_Delete_False,
	}).GroupBy("edit_user_id")
	rs, err = db.Query(sql)
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

// 根据文档信息获取文档内容和文件地址
func (d *Document) GetDocumentContentByDocument(doc map[string]string) (content string, pageFile string, err error) {
	// get document page file
	_, pageFile, err = DocumentModel.GetParentDocumentsByDocument(doc)
	if err != nil {
		return content, pageFile, err
	}
	// get document content
	content, err = utils.Document.GetContentByPageFile(pageFile)
	if err != nil {
		return content, pageFile, err
	}
	return content, pageFile, nil
}

// ==================== 回收站相关方法 ====================

// 软删除文档（进入回收站）：标记 is_delete=1, deleted_time=now，不删文件
func (d *Document) SoftDelete(documentId string, spaceId string, userId string) (err error) {
	db := G.DB()
	_, err = db.Exec(db.AR().Update(Table_Document_Name, map[string]interface{}{
		"is_delete":    Document_Delete_True,
		"deleted_time": time.Now().Unix(),
		"update_time":  time.Now().Unix(),
		"edit_user_id": userId,
	}, map[string]interface{}{
		"document_id": documentId,
	}))
	if err != nil {
		return
	}

	// create document log
	go func(userId, documentId, spaceId string) {
		_, err := LogDocumentModel.DeleteAction(userId, documentId, spaceId)
		if err != nil {
			logs.Error("soft delete document add log err=%s", err.Error())
		}
	}(userId, documentId, spaceId)

	return
}

// 恢复文档（从回收站还原）：is_delete=0, deleted_time=0
func (d *Document) RecoverDocument(documentId string) (err error) {
	db := G.DB()
	_, err = db.Exec(db.AR().Update(Table_Document_Name, map[string]interface{}{
		"is_delete":    Document_Delete_False,
		"deleted_time": 0,
		"update_time":  time.Now().Unix(),
	}, map[string]interface{}{
		"document_id": documentId,
	}))
	return
}

// 彻底删除文档（物理删文件 + 删数据库记录）
func (d *Document) PermanentlyDelete(documentId string) (err error) {
	document, err := d.GetDeletedDocumentByDocumentId(documentId)
	if err != nil || len(document) == 0 {
		return
	}

	spaceId := document["space_id"]
	userId := document["edit_user_id"]

	// 获取文件路径（用于物理删除）
	// 回收站中的文档，父文档可能已被删，用 GetParentDocumentsByDocument 可能失败
	// 所以用原始路径信息直接计算
	pageFile, fileErr := d.GetPageFileByDocument(document)
	if fileErr != nil {
		logs.Error("get deleted document page file err: %s", fileErr.Error())
	}

	// 物理删除文件
	if pageFile != "" {
		_ = utils.Document.Delete(pageFile, utils.Convert.StringToInt(document["type"]))
	}

	// 删除附件文件和数据库记录
	_ = AttachmentModel.DeleteAttachmentsDBFileByDocumentId(documentId)

	// 删除数据库记录
	db := G.DB()
	_, err = db.Exec(db.AR().Delete(Table_Document_Name, map[string]interface{}{
		"document_id": documentId,
	}))
	if err != nil {
		return
	}

	// delete follow doc
	go func(documentId string) {
		_ = FollowModel.DeleteByObjectIdType(documentId, fmt.Sprintf("%d", Follow_Type_Doc))
	}(documentId)

	// delete collect doc
	go func(documentId string) {
		_ = CollectionModel.DeleteByResourceIdType(documentId, fmt.Sprintf("%d", Collection_Type_Doc))
	}(documentId)

	// create document log
	go func(userId, documentId, spaceId string) {
		_, _ = LogDocumentModel.DeleteAction(userId, documentId, spaceId)
	}(userId, documentId, spaceId)

	return
}

// 获取回收站中的文档（通过 document_id，不限制 is_delete）
func (d *Document) GetDeletedDocumentByDocumentId(documentId string) (document map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"document_id":  documentId,
		"is_delete":    Document_Delete_True,
		"deleted_time >": 0,
	}))
	if err != nil {
		return
	}
	document = rs.Row()
	return
}

// 获取某空间下回收站中的文档列表（分页）
func (d *Document) GetDeletedDocumentsBySpaceId(spaceId string, limit int, number int) (documents []map[string]string, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
		"space_id":      spaceId,
		"is_delete":     Document_Delete_True,
		"deleted_time >": 0,
		"parent_id >":   "0",
	}).OrderBy("deleted_time", "DESC").Limit(limit, number))
	if err != nil {
		return
	}
	documents = rs.Rows()
	return
}

// 统计某空间下回收站文档数量
func (d *Document) CountDeletedDocumentsBySpaceId(spaceId string) (count int64, err error) {
	db := G.DB()
	var rs *mysql.ResultSet
	rs, err = db.Query(
		db.AR().
			Select("count(*) as total").
			From(Table_Document_Name).
			Where(map[string]interface{}{
				"space_id":      spaceId,
				"is_delete":     Document_Delete_True,
				"deleted_time >": 0,
				"parent_id >":   "0",
			}))
	if err != nil {
		return
	}
	count = utils.NewConvert().StringToInt64(rs.Value("total"))
	return
}

// 获取所有已过期需要彻底删除的回收站文档
// 条件：is_delete=1 AND deleted_time > 0 AND (now - deleted_time) > space.recycle_keep_days * 86400
func (d *Document) GetExpiredDeletedDocuments() (documents []map[string]string, err error) {
	documents = []map[string]string{}

	// 获取所有空间
	spaces, err := SpaceModel.GetSpaces()
	if err != nil {
		return
	}

	now := time.Now().Unix()

	for _, space := range spaces {
		keepDays := utils.Convert.StringToInt64(space["recycle_keep_days"])
		// 只处理 keepDays > 0 的情况（有保留期限的空间）
		if keepDays <= 0 {
			continue
		}

		expireTime := now - keepDays*86400
		// 查询该空间下已过期的回收站文档
		db := G.DB()
		var rs *mysql.ResultSet
		rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
			"space_id":        space["space_id"],
			"is_delete":       Document_Delete_True,
			"deleted_time >":  0,
			"deleted_time <":  expireTime,
			"parent_id >":     "0",
		}))
		if err != nil {
			logs.Error("get expired documents for space %s err: %s", space["space_id"], err.Error())
			continue
		}
		documents = append(documents, rs.Rows()...)
	}
	return
}

// 获取空间下 recycle_keep_days=0 且已软删除的文档（应立即删除的）
func (d *Document) GetImmediatelyDeleteDocuments() (documents []map[string]string, err error) {
	documents = []map[string]string{}

	// 获取所有 recycle_keep_days=0 的空间
	spaces, err := SpaceModel.GetSpaces()
	if err != nil {
		return
	}

	for _, space := range spaces {
		keepDays := utils.Convert.StringToInt64(space["recycle_keep_days"])
		if keepDays != 0 {
			continue
		}

		// 查询该空间下已软删除的文档
		db := G.DB()
		var rs *mysql.ResultSet
		rs, err = db.Query(db.AR().From(Table_Document_Name).Where(map[string]interface{}{
			"space_id":       space["space_id"],
			"is_delete":      Document_Delete_True,
			"deleted_time >": 0,
			"parent_id >":    "0",
		}))
		if err != nil {
			logs.Error("get immediately delete documents for space %s err: %s", space["space_id"], err.Error())
			continue
		}
		documents = append(documents, rs.Rows()...)
	}
	return
}

// 根据文档的 path + name + type 计算文件路径（用于回收站文档，避免查父文档）
func (d *Document) GetPageFileByDocument(document map[string]string) (pageFile string, err error) {
	if document["parent_id"] == "0" {
		pageFile = utils.Document.GetDefaultPageFileBySpaceName(document["name"])
	} else {
		// path 存的是 "0,1,2" 这样的父文档ID链
		// 回收站中父文档可能也被删了，用 GetAllDocumentsByDocumentIds 获取（不过滤 is_delete）
		pathIds := strings.Split(document["path"], ",")
		parentDocuments, pErr := d.GetAllDocumentsByDocumentIds(pathIds)
		if pErr != nil {
			err = pErr
			return
		}
		var parentPath = ""
		for _, parentDocument := range parentDocuments {
			parentPath += parentDocument["name"] + "/"
		}
		parentPath = strings.TrimRight(parentPath, "/")
		pageFile = utils.Document.GetPageFileByParentPath(document["name"], utils.Convert.StringToInt(document["type"]), parentPath)
	}
	return
}
