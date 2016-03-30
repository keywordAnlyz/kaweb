package routers

import (
	"github.com/keywordAnlyz/kaweb/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
}
