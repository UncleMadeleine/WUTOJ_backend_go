package controller

import (
	"OnlineJudge/app/api/model"
	"OnlineJudge/app/common/validate"
	"OnlineJudge/app/helper"
	"OnlineJudge/constants"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func GetAllProblems(c *gin.Context) {

	problemModel := model.Problem{}
	problemJson := struct {
		PageNumber int `json:"page_number" form:"page_number"`
	}{}
	if err := c.ShouldBind(&problemJson);err != nil {
		c.JSON(http.StatusOK, helper.ApiReturn(constants.CodeError, "页码参数错误", err.Error()))
		return
	}
	res := problemModel.GetAllProblems((problemJson.PageNumber-1)*constants.PageLimit,constants.PageLimit)
	c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
	return

}

func GetProblemByID(c *gin.Context) {
	problemValidate := validate.ProblemValidate
	problemModel := model.Problem{}
	contestModel := model.Contest{}
	contestUserModel := model.ContestUser{}

	problemJson := struct {
		ProblemID int `json:"problem_id"`
	}{}
	if problemID,err := strconv.Atoi(c.Param("problem_id"));err != nil{
		c.JSON(http.StatusOK, helper.ApiReturn(constants.CodeError, "参数错误", err.Error()))
		return
	}else{
		problemJson.ProblemID = problemID
	}
	if err := c.ShouldBind(&problemJson); err != nil {
		c.JSON(http.StatusOK, helper.ApiReturn(constants.CodeError, "绑定数据模型失败", err.Error()))
		return
	}

	problemMap := helper.Struct2Map(problemJson)
	if res, err := problemValidate.ValidateMap(problemMap, "findByID"); !res {
		c.JSON(http.StatusOK, helper.ApiReturn(constants.CodeError, err.Error(), 0))
		return
	}

	// TODO: need remove, temprory workaround
	contestJson := contestModel.GetContestByProblemId(problemMap["problem_id"].(int))
	if contestJson.Status == constants.CodeSuccess {
		userJson := checkLogin(c)
		if userJson.Status != constants.CodeSuccess{
			c.JSON(http.StatusOK, helper.ApiReturn(constants.CodeError, "未登录", 0))
			return
		}
		userIDRaw := userJson.Data.(uint)
		contest := contestJson.Data.(model.Contest)
		userID := int(userIDRaw)
		contestID := contest.ContestID
		if participation := contestUserModel.CheckUserContest(userID,contestID); participation.Status != constants.CodeSuccess{
			c.JSON(http.StatusOK, helper.ApiReturn(constants.CodeError, "尚未参赛，请参赛", 0))
			return
		}

		format := "2006-01-02 15:04:05"
		now, _ := time.Parse(format, time.Now().Format(format))
		if now.Before(contest.BeginTime) || contest.EndTime.Before(now) {
			c.JSON(http.StatusOK, helper.ApiReturn(constants.CodeError, "比赛未开始", 0))
			return
		}
	}

	res := problemModel.GetProblemByID(int(problemJson.ProblemID))
	c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
	return
}

func SearchProblem(c *gin.Context) {
	problemJson := struct {
		Param string `uri:"param" json:"param"`
	}{}

	problemJson.Param = c.Param("param")
	problemModel := model.Problem{}
	contestModel := model.Contest{}

	if err := c.ShouldBind(&problemJson); err == nil {
		res := problemModel.SearchProblem(problemJson.Param)

		// TODO: need remove, temprory workaround
		problem_id, _ := strconv.Atoi(problemJson.Param)
		contestJson := contestModel.GetContestByProblemId(problem_id)
		if contestJson.Status == constants.CodeSuccess {
			format := "2006-01-02 15:04:05"
			now, _ := time.Parse(format, time.Now().Format(format))
			contest := contestJson.Data.(model.Contest)
			if now.Before(contest.BeginTime) || contest.EndTime.Before(now) {
				c.JSON(http.StatusOK, helper.ApiReturn(constants.CodeError, "比赛未开始", 0))
				return
			}
		}
		c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
		return
	} else {
		c.JSON(http.StatusOK, helper.ApiReturn(constants.CodeError, "数据模型绑定错误", err.Error()))
		return
	}
}
