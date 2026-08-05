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

// 回收站主页面（左侧空间列表 + 右侧 iframe）
func (this *RecycleController) Index() {

	spaceId := this.GetString("space_id", "")

	// 获取所有空间供选择
	spaces, err := models.SpaceModel.GetSpaces()
	if err != nil {
		this.ErrorLog("获取回收站空间列表失败: " + err.Error())
		this.ViewError("获取回收站失败！", "/main/index")
	}

	this.Data["spaces"] = spaces
	this.Data["current_space_id"] = spaceId

	this.viewLayout("recycle/index", "space")
}

// 回收站列表（iframe 内加载）
func (this *RecycleController) List() {

	spaceId := this.GetString("space_id", "")

	if spaceId == "" {
		this.Data["documents"] = []map[string]string{}
		this.Data["space"] = map[string]string{}
		this.Data["count"] = int64(0)
		this.viewLayout("recycle/list", "default")
		return
	}

	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("获取空间失败: " + err.Error())
		this.ViewError("空间不存在！", "/recycle/index")
	}
	if len(space) == 0 {
		this.ViewError("空间不存在！", "/recycle/index")
	}

	// 检查权限
	isVisit, _, _ := this.GetDocumentPrivilege(space)
	if !isVisit {
		this.ViewError("您没有权限查看该空间回收站！")
	}

	page, _ := this.GetInt("page", 1)
	number, _ := this.GetRangeInt("number", 20, 10, 100)
	limit := (page - 1) * number

	count, err := models.DocumentModel.CountDeletedDocumentsBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("获取回收站文档数量失败: " + err.Error())
		this.ViewError("获取回收站失败！", "/recycle/index")
	}

	documents, err := models.DocumentModel.GetDeletedDocumentsBySpaceId(spaceId, limit, number)
	if err != nil {
		this.ErrorLog("获取回收站文档列表失败: " + err.Error())
		this.ViewError("获取回收站失败！", "/recycle/index")
	}

	// 获取文档创建者和删除者信息
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
		// 计算剩余天数
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
	this.SetPaginator(number, count)

	this.viewLayout("recycle/list", "default")
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
	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("恢复文档失败：" + err.Error())
		this.jsonError("恢复文档失败！")
	}
	if len(space) == 0 {
		this.jsonError("文档所在空间不存在！")
	}

	// 检查权限
	_, _, isManager := this.GetDocumentPrivilege(space)
	if !isManager {
		this.jsonError("您没有权限恢复该文档！")
	}

	err = models.DocumentModel.RecoverDocument(documentId)
	if err != nil {
		this.ErrorLog("恢复文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("恢复文档失败！")
	}

	this.InfoLog("恢复文档 " + documentId + " 成功")
	this.jsonSuccess("恢复文档成功", nil, "/recycle/list?space_id="+spaceId)
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
	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("彻底删除文档失败：" + err.Error())
		this.jsonError("彻底删除文档失败！")
	}
	if len(space) == 0 {
		this.jsonError("文档所在空间不存在！")
	}

	// 检查权限
	_, _, isManager := this.GetDocumentPrivilege(space)
	if !isManager {
		this.jsonError("您没有权限彻底删除该文档！")
	}

	err = models.DocumentModel.PermanentlyDelete(documentId)
	if err != nil {
		this.ErrorLog("彻底删除文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("彻底删除文档失败！")
	}

	this.InfoLog("彻底删除文档 " + documentId + " 成功")
	this.jsonSuccess("彻底删除成功", nil, "/recycle/list?space_id="+spaceId)
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

	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("清空回收站失败：" + err.Error())
		this.jsonError("清空回收站失败！")
	}
	if len(space) == 0 {
		this.jsonError("空间不存在！")
	}

	// 检查权限
	_, _, isManager := this.GetDocumentPrivilege(space)
	if !isManager {
		this.jsonError("您没有权限清空该空间回收站！")
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

	this.InfoLog(fmt.Sprintf("清空回收站 成功%d 失败%d", successCount, failCount))
	this.jsonSuccess(fmt.Sprintf("清空完成！成功 %d 个，失败 %d 个", successCount, failCount), nil, "/recycle/list?space_id="+spaceId)
}
