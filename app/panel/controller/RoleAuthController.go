package controller

import (
	"OnlineJudge/app/common/validate"
	"OnlineJudge/app/helper"
	"OnlineJudge/app/panel/model"
	"OnlineJudge/constants"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetRoleAuthsList(c *gin.Context) {
	roleAuthValidate := validate.RoleAuthValidate //jun
	authModel := model.Auth{}

	authJson := struct {
		Rid int `json:"rid" form:"rid"`
	}{}

	if err := c.ShouldBind(&authJson); err != nil {
		c.JSON(http.StatusOK, helper.BackendApiReturn(constants.CodeError, "绑定数据模型失败", err.Error()))
		return
	}

	authMap := helper.Struct2Map(authJson)
	if res, err := roleAuthValidate.ValidateMap(authMap, "getRoleAuth"); !res {
		c.JSON(http.StatusOK, helper.BackendApiReturn(constants.CodeError, err.Error(), 0))
		return
	}

	allAuths := authModel.GetAuthNoRules()

	res := authModel.GetRoleAuth(authJson.Rid)
	if res.Status != constants.CodeSuccess {
		c.JSON(http.StatusOK, helper.BackendApiReturn(res.Status, res.Msg, res.Data))
		return
	}
	auths := res.Data.([]model.Auth)
	var val []int
	for _, auth := range auths {
		val = append(val, auth.Aid)
	}
	c.JSON(http.StatusOK, helper.BackendApiReturn(res.Status, res.Msg, map[string]interface{}{
		"allAuths": allAuths.Data,
		"values":   val,
	}))
	return
}

func AddRoleAuths(c *gin.Context) {
	roleAuthValidate := validate.RoleAuthValidate
	roleAuthModel := model.RoleAuth{}

	roleAuthsJson := struct {
		Rid  int    `json:"rid" form:"rid"`
		Aids string `json:"aids" form:"aids"`
	}{}

	if err := c.ShouldBind(&roleAuthsJson); err != nil {
		c.JSON(http.StatusOK, helper.BackendApiReturn(constants.CodeError, "绑定数据模型失败", err.Error()))
		return
	}

	roleAuthsMap := helper.Struct2Map(roleAuthsJson)
	if res, err := roleAuthValidate.ValidateMap(roleAuthsMap, "addGroup"); !res {
		c.JSON(http.StatusOK, helper.BackendApiReturn(constants.CodeError, err.Error(), 0))
		return
	}

	var aids []int
	_ = json.Unmarshal([]byte((roleAuthsJson.Aids)), &aids)
	fmt.Println(aids)
	for _, aid := range aids {
		res := roleAuthModel.AddRoleAuth(model.RoleAuth{Rid: roleAuthsJson.Rid, Aid: aid})
		if res.Status != constants.CodeSuccess {
			c.JSON(http.StatusOK, helper.BackendApiReturn(res.Status, "编号为"+string(rune(aid))+"的权限添加失败", res.Data))
			return
		}
	}
	c.JSON(http.StatusOK, helper.BackendApiReturn(constants.CodeSuccess, "添加成功", true))
	return
}

func DeleteRoleAuths(c *gin.Context) {
	roleAuthValidate := validate.RoleAuthValidate
	roleAuthModel := model.RoleAuth{}

	roleAuthsJson := struct {
		Rid  int    `json:"rid" form:"rid"`
		Aids string `json:"aids" form:"aids"`
	}{}

	if err := c.ShouldBind(&roleAuthsJson); err != nil {
		c.JSON(http.StatusOK, helper.BackendApiReturn(constants.CodeError, "绑定数据模型失败", err.Error()))
		return
	}

	roleAuthsMap := helper.Struct2Map(roleAuthsJson)
	if res, err := roleAuthValidate.ValidateMap(roleAuthsMap, "deleteGroup"); !res {
		c.JSON(http.StatusOK, helper.BackendApiReturn(constants.CodeError, err.Error(), 0))
		return
	}

	var aids []int
	_ = json.Unmarshal([]byte((roleAuthsJson.Aids)), &aids)
	fmt.Println(aids)
	for _, aid := range aids {
		res := roleAuthModel.DeleteRoleAuth(model.RoleAuth{Rid: roleAuthsJson.Rid, Aid: aid})
		if res.Status != constants.CodeSuccess {
			c.JSON(http.StatusOK, helper.BackendApiReturn(res.Status, "编号为"+string(rune(aid))+"的权限添加失败", res.Data))
			return
		}
	}
	c.JSON(http.StatusOK, helper.BackendApiReturn(constants.CodeSuccess, "添加成功", true))
	return
}
