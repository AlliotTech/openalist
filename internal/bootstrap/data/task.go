package data

import (
	"github.com/AlliotTech/openalist/internal/db"
	"github.com/AlliotTech/openalist/internal/model"
)

var initialTaskItems []model.TaskItem

func initTasks() {
	InitialTasks()

	for i := range initialTaskItems {
		item := &initialTaskItems[i]
		taskitem, _ := db.GetTaskDataByType(item.Key)
		if taskitem == nil {
			db.CreateTaskData(item)
		}
	}
}

func InitialTasks() []model.TaskItem {
	initialTaskItems = []model.TaskItem{
		{Key: "copy", PersistData: "[]"},
		{Key: "download", PersistData: "[]"},
		{Key: "transfer", PersistData: "[]"},
	}
	return initialTaskItems
}
