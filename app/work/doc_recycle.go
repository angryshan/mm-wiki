package work

import (
	"github.com/astaxie/beego/logs"
	"github.com/phachon/mm-wiki/app/models"
	"sync"
	"time"
)

var (
	DocRecycleWorker = NewDocRecycleWork()
)

type DocRecycle struct {
	lock          sync.RWMutex
	runStatus     int
	isTaskRunning bool
	quit          chan bool
}

func NewDocRecycleWork() *DocRecycle {
	return &DocRecycle{
		runStatus:     RunStatusStop,
		isTaskRunning: false,
		quit:          make(chan bool, 1),
	}
}

// Start 开始定时清理回收站
func (d *DocRecycle) Start() {
	if d.runStatus == RunStatusRunning {
		return
	}
	// 每隔 1 小时执行一次
	d.cleanExpiredDocuments()
	go func(d *DocRecycle) {
		defer func() {
			e := recover()
			if e != nil {
				logs.Info("[DocRecycleWork] clean panic: %v", e)
			}
			d.lock.Lock()
			d.runStatus = RunStatusStop
			d.isTaskRunning = false
			d.lock.Unlock()
		}()
		d.lock.Lock()
		d.runStatus = RunStatusRunning
		d.lock.Unlock()
		for {
			select {
			case <-time.Tick(1 * time.Hour):
				if !d.isTaskRunning {
					d.cleanExpiredDocuments()
				}
			case <-d.quit:
				logs.Info("[DocRecycleWork] stop recycle worker")
				return
			}
		}
	}(d)
}

// Stop 停止
func (d *DocRecycle) Stop() {
	d.quit <- true
}

// 清理过期的回收站文档
func (d *DocRecycle) cleanExpiredDocuments() {

	logs.Info("[DocRecycleWork] start clean expired documents")

	d.lock.Lock()
	d.isTaskRunning = true
	d.lock.Unlock()

	// 1. 清理 recycle_keep_days=0 的空间（应立即删除）
	immediateDocs, err := models.DocumentModel.GetImmediatelyDeleteDocuments()
	if err != nil {
		logs.Error("[DocRecycleWork] get immediately delete documents err: %s", err.Error())
	} else {
		for _, doc := range immediateDocs {
			err := models.DocumentModel.PermanentlyDelete(doc["document_id"])
			if err != nil {
				logs.Error("[DocRecycleWork] permanently delete document %s err: %s", doc["document_id"], err.Error())
			} else {
				logs.Info("[DocRecycleWork] permanently deleted document %s (immediate)", doc["document_id"])
			}
		}
	}

	// 2. 清理已过期的回收站文档
	expiredDocs, err := models.DocumentModel.GetExpiredDeletedDocuments()
	if err != nil {
		logs.Error("[DocRecycleWork] get expired deleted documents err: %s", err.Error())
	} else {
		for _, doc := range expiredDocs {
			err := models.DocumentModel.PermanentlyDelete(doc["document_id"])
			if err != nil {
				logs.Error("[DocRecycleWork] permanently delete document %s err: %s", doc["document_id"], err.Error())
			} else {
				logs.Info("[DocRecycleWork] permanently deleted document %s (expired)", doc["document_id"])
			}
		}
	}

	d.lock.Lock()
	d.isTaskRunning = false
	d.lock.Unlock()

	logs.Info("[DocRecycleWork] finish clean expired documents, immediate=%d, expired=%d",
		len(immediateDocs), len(expiredDocs))
}
