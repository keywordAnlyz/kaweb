package routers

import (
	"github.com/astaxie/beego"
	"github.com/keywordAnlyz/kaweb/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/index.html", &controllers.MainController{})
	beego.Router("/global/list.html", &controllers.GlobalController{}, "get:List")
	beego.Router("/global/item/update", &controllers.GlobalController{}, "post:UpdateItem")

	beego.Router("/task/add.html", &controllers.TaskController{}, "get,post:Upload")
	beego.Router("/task/list.html", &controllers.TaskController{}, "get,post:List")
	beego.Router("/task/:id:int/detail.html", &controllers.TaskController{}, "get:Detail")
	beego.Router("/task/:id:int/start.html", &controllers.TaskController{}, "post,get:StartTask")
	beego.Router("/task/:id:int/words.html", &controllers.TaskController{}, "post,get:ShowWords")

	beego.Router("/task/:taskId:int/word/:wordId:int/detail.html", &controllers.TaskController{}, "post,get:WordDetail")

	beego.Router("/task/:taskId:int/word/:wordId:int/view.html", &controllers.FileViewController{}, "post,get:Get")

	beego.Router("/report/show.html", &controllers.ReportController{}, "get,post:Get")

}
