package controllers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/phachon/mm-wiki/app/models"
	"github.com/phachon/mm-wiki/app/utils"
)

type RecycleController struct {
	BaseController
}

// 回收站列表页面（iframe 内加载）
func (this *RecycleController) List() {

	spaceId := this.GetString("space_id", "")

	// 获取所有空间供选择
	spaces, err := models.SpaceModel.GetSpaces()
	if err != nil {
		this.ErrorLog("获取空间列表失败: " + err.Error())
		this.ViewError("获取空间列表失败！", "/system/main/index")
	}
	this.Data["spaces"] = spaces
	this.Data["current_space_id"] = spaceId

	if spaceId == "" {
		this.Data["documents"] = []map[string]string{}
		this.Data["space"] = map[string]string{}
		this.Data["count"] = int64(0)
		this.Data["is_manager"] = false
		this.viewLayout("recycle/list", "default")
		return
	}

	var space map[string]string
	space, err = models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("获取空间失败: " + err.Error())
		this.ViewError("空间不存在！", "/system/main/index")
	}
	if len(space) == 0 {
		this.ViewError("空间不存在！", "/system/main/index")
	}

	page, _ := this.GetInt("page", 1)
	number, _ := this.GetRangeInt("number", 20, 10, 100)
	limit := (page - 1) * number

	count, err := models.DocumentModel.CountDeletedDocumentsBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("获取回收站文档数量失败: " + err.Error())
		this.ViewError("获取回收站失败！", "/system/main/index")
	}

	documents, err := models.DocumentModel.GetDeletedDocumentsBySpaceId(spaceId, limit, number)
	if err != nil {
		this.ErrorLog("获取回收站文档列表失败: " + err.Error())
		this.ViewError("获取回收站失败！", "/system/main/index")
	}

	// 获取文档删除者信息
	userIds := []string{}
	for _, doc := range documents {
		if doc["edit_user_id"] != "" {
			userIds = append(userIds, doc["edit_user_id"])
		}
		if doc["create_user_id"] != "" {
			userIds = append(userIds, doc["create_user_id"])
		}
	}
	users, _ := models.UserModel.GetUsersByUserIds(userIds)
	userMap := map[string]string{}
	for _, user := range users {
		userMap[user["user_id"]] = user["username"]
	}

	// 计算剩余天数
	now := time.Now().Unix()
	for _, doc := range documents {
		doc["delete_username"] = userMap[doc["edit_user_id"]]
		doc["create_username"] = userMap[doc["create_user_id"]]
		deletedTime := utils.Convert.StringToInt64(doc["deleted_time"])
		keepDays := utils.Convert.StringToInt64(space["recycle_keep_days"])
		expireTime := deletedTime + keepDays*86400
		remainSeconds := expireTime - now
		if remainSeconds <= 0 {
			doc["remain_days"] = "0"
		} else {
			doc["remain_days"] = strconv.FormatInt(remainSeconds/86400+1, 10)
		}
	}

	this.Data["documents"] = documents
	this.Data["space"] = space
	this.Data["count"] = count
	this.Data["is_manager"] = true // 系统模块下默认有管理权限
	this.SetPaginator(number, count)

	this.viewLayout("recycle/list", "default")
}

// 查看回收站文档内容（新页面打开，复用分享页样式）
func (this *RecycleController) View() {

	documentId := this.GetString("document_id", "")
	if documentId == "" {
		this.ViewError("文档未找到！", "/system/main/index")
	}

	document, err := models.DocumentModel.GetDeletedDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("查找回收站文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("查找文档失败！", "/system/main/index")
	}
	if len(document) == 0 {
		this.ViewError("文档不存在或不在回收站中！", "/system/main/index")
	}

	// 获取父文档路径
	parentDocuments, pageFile, err := models.DocumentModel.GetParentDocumentsByDocument(document)
	if err != nil {
		this.ErrorLog("查找父文档失败：" + err.Error())
		this.ViewError("查找文档失败！", "/system/main/index")
	}

	// 获取文档内容
	documentContent := ""
	if document["type"] != "2" {
		documentContent, err = utils.Document.GetContentByPageFile(pageFile)
		if err != nil {
			this.ErrorLog("查找文档 " + documentId + " 内容失败：" + err.Error())
			documentContent = ""
		}
	}

	// 获取创建者和修改者
	users, _ := models.UserModel.GetUsersByUserIds([]string{document["create_user_id"], document["edit_user_id"]})
	var createUser = map[string]string{}
	var editUser = map[string]string{}
	for _, user := range users {
		if user["user_id"] == document["create_user_id"] {
			createUser = user
		}
		if user["user_id"] == document["edit_user_id"] {
			editUser = user
		}
	}

	this.Data["create_user"] = createUser
	this.Data["edit_user"] = editUser
	this.Data["document"] = document
	this.Data["page_content"] = documentContent
	this.Data["parent_documents"] = parentDocuments

	// 直接复用全局的 page/display 视图和 document_share 布局
	this.ViewLayout("page/display", "layouts/document_share")
}

// 恢复文档
func (this *RecycleController) Recover() {

	if !this.IsPost() {
		this.jsonError("请求方式有误！")
	}

	documentId := this.GetString("document_id", "0")
	if documentId == "0" {
		this.jsonError("没有选择文档！")
	}

	document, err := models.DocumentModel.GetDeletedDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("恢复文档失败：" + err.Error())
		this.jsonError("恢复文档失败！")
	}
	if len(document) == 0 {
		this.jsonError("文档不存在或不在回收站中！")
	}

	spaceId := document["space_id"]

	err = models.DocumentModel.RecoverDocument(documentId)
	if err != nil {
		this.ErrorLog("恢复文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("恢复文档失败！")
	}

	// 记录恢复文档日志
	go func(userId, docId, spcId string) {
		_, _ = models.LogDocumentModel.RecoverAction(userId, docId, spcId)
	}(this.UserId, documentId, spaceId)

	this.InfoLog("恢复文档 " + documentId + " 成功")
	this.jsonSuccess("恢复文档成功", nil, "/system/recycle/list?space_id="+spaceId)
}

// 彻底删除文档
func (this *RecycleController) Remove() {

	if !this.IsPost() {
		this.jsonError("请求方式有误！")
	}

	documentId := this.GetString("document_id", "0")
	if documentId == "0" {
		this.jsonError("没有选择文档！")
	}

	document, err := models.DocumentModel.GetDeletedDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("彻底删除文档失败：" + err.Error())
		this.jsonError("彻底删除文档失败！")
	}
	if len(document) == 0 {
		this.jsonError("文档不存在或不在回收站中！")
	}

	spaceId := document["space_id"]

	err = models.DocumentModel.PermanentlyDelete(documentId)
	if err != nil {
		this.ErrorLog("彻底删除文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("彻底删除文档失败！")
	}

	// 记录彻底删除文档日志
	go func(userId, docId, spcId string) {
		_, _ = models.LogDocumentModel.PermanentlyDeleteAction(userId, docId, spcId)
	}(this.UserId, documentId, spaceId)

	this.InfoLog("彻底删除文档 " + documentId + " 成功")
	this.jsonSuccess("彻底删除成功", nil, "/system/recycle/list?space_id="+spaceId)
}

// 清空某空间回收站
func (this *RecycleController) Clear() {

	if !this.IsPost() {
		this.jsonError("请求方式有误！")
	}

	spaceId := this.GetString("space_id", "")
	if spaceId == "" {
		this.jsonError("没有选择空间！")
	}

	// 获取该空间所有回收站文档
	documents, err := models.DocumentModel.GetDeletedDocumentsBySpaceId(spaceId, 0, 10000)
	if err != nil {
		this.ErrorLog("清空回收站失败：" + err.Error())
		this.jsonError("清空回收站失败！")
	}

	successCount := 0
	failCount := 0
	for _, doc := range documents {
		err := models.DocumentModel.PermanentlyDelete(doc["document_id"])
		if err != nil {
			failCount++
		} else {
			successCount++
		}
	}

	// 记录清空回收站日志
	go func(userId, spcId string, count int) {
		_, _ = models.LogDocumentModel.ClearRecycleAction(userId, spcId, count)
	}(this.UserId, spaceId, successCount)

	this.InfoLog(fmt.Sprintf("清空回收站 成功%d 失败%d", successCount, failCount))
	this.jsonSuccess(fmt.Sprintf("清空完成！成功 %d 个，失败 %d 个", successCount, failCount), nil, "/system/recycle/list?space_id="+spaceId)
}
