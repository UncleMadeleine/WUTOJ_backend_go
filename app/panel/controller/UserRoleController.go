package controller

import (
	"OnlineJudge/app/common"
	"OnlineJudge/app/common/validate"
	"OnlineJudge/app/helper"
	"OnlineJudge/app/panel/model"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// TODO
func AddUserRoles(c *gin.Context) {
	if res := haveAuth(c, "roleAssign"); res != common.Authed {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "权限不足", res))
		return
	}
	userRoleValidate := validate.UserRoleValidate
	userRoleModel := model.UserRole{}

	userRolesJson := struct {
		UserID int    `json:"user_id" form:"user_id"`
		Rids   string `json:"rids" form:"rids"`
	}{}

	if err := c.ShouldBind(&userRolesJson); err != nil {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "绑定数据模型失败", err.Error()))
		return
	}

	userRolesMap := helper.Struct2Map(userRolesJson)
	if res, err := userRoleValidate.ValidateMap(userRolesMap, "addGroup"); !res {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, err.Error(), 0))
		return
	}

	var rids []int
	_ = json.Unmarshal([]byte((userRolesJson.Rids)), &rids)
	fmt.Println(rids)
	for _, rid := range rids {
		res := userRoleModel.AddUserRole(model.UserRole{UserID: userRolesJson.UserID, Rid: rid})
		if res.Status != common.CodeSuccess {
			c.JSON(http.StatusOK, helper.BackendApiReturn(res.Status, "编号为"+string(rune(rid))+"的角色添加失败", res.Data))
			return
		}
	}
	c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeSuccess, "添加成功", true))
	return
}

func DeleteUserRoles(c *gin.Context) {
	if res := haveAuth(c, "roleAssign"); res != common.Authed {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "权限不足", res))
		return
	}
	userRoleValidate := validate.UserRoleValidate
	userRoleModel := model.UserRole{}

	userRolesJson := struct {
		UserID int    `json:"user_id" form:"user_id"`
		Rids   string `json:"rids" form:"rids"`
	}{}

	if err := c.ShouldBind(&userRolesJson); err != nil {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "绑定数据模型失败", false))
		return
	}

	userRolesMap := helper.Struct2Map(userRolesJson)
	if res, err := userRoleValidate.ValidateMap(userRolesMap, "deleteGroup"); !res {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, err.Error(), 0))
		return
	}

	var rids []int
	_ = json.Unmarshal([]byte((userRolesJson.Rids)), &rids)
	fmt.Println(rids)
	for _, rid := range rids {
		res := userRoleModel.DeleteUserRole(model.UserRole{UserID: userRolesJson.UserID, Rid: rid})
		if res.Status != common.CodeSuccess {
			c.JSON(http.StatusOK, helper.BackendApiReturn(res.Status, "编号为"+string(rune(rid))+"的权限删除失败", res.Data))
			return
		}
	}
	c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeSuccess, "删除成功", true))
	return

}

func GetUserRolesList(c *gin.Context) {
	if res := haveAuth(c, "roleAssign"); res != common.Authed {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "权限不足", res))
		return
	}
	userRoleValidate := validate.UserRoleValidate
	roleModel := model.Role{}

	roleJson := struct {
		UserID int `json:"user_id" form:"user_id"`
	}{}

	if err := c.ShouldBind(&roleJson); err != nil {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "绑定数据模型失败", err.Error()))
		return
	}

	roleMap := helper.Struct2Map(roleJson)
	if res, err := userRoleValidate.ValidateMap(roleMap, "getUserRole"); !res {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, err.Error(), 0))
		return
	}

	allRoles := roleModel.GetRoleNoRules()

	res := roleModel.GetUserRole(roleJson.UserID)
	roles := res.Data.([]model.Role)
	var val []int
	for _, role := range roles {
		val = append(val, role.Rid)
	}
	c.JSON(http.StatusOK, helper.BackendApiReturn(res.Status, res.Msg, map[string]interface{}{
		"allRoles": allRoles.Data,
		"values":   val,
	}))
	return
}
