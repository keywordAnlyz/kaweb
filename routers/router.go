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

}
