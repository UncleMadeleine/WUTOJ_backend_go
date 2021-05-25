package controller

import (
	"OnlineJudge/app/common"
	"OnlineJudge/app/helper"
	"OnlineJudge/app/panel/model"
	"OnlineJudge/db_server"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetContestBalloon(c *gin.Context)  {
	if res := haveAuth(c, "getBalloonStatus"); res != common.Authed {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "权限不足", res))
		return
	}

	contestModel := model.Contest{}
	SubmitModel := model.Submit{}

	contestIDJson := struct {
		ContestID 	uint 	`json:"contest_id" form:"contest_id"`
	}{}

	if err := c.ShouldBind(&contestIDJson); err != nil {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "绑定数据模型失败", err.Error()))
		return
	}

	contestRes := contestModel.GetContestById(contestIDJson.ContestID)
	if contestRes.Status != common.CodeSuccess {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, contestRes.Msg, contestRes.Data))
		return
	}

	submitRes := SubmitModel.GetContestACSubmits(contestIDJson.ContestID)
	if submitRes.Status != common.CodeSuccess {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, submitRes.Msg, submitRes.Data))
		return
	}

	contest := contestRes.Data.(model.Contest)
	submits := submitRes.Data.([]model.Submit)

	colors := make([]string, 0)
	problems := make([]uint, 0)
	_ = json.Unmarshal([]byte(contest.Colors), &colors)
	_ = json.Unmarshal([]byte(contest.Problems), &problems)
	fmt.Fprintf(gin.DefaultWriter, "colors: %v\n problems: %v\n", colors, problems)

	problemIDMap := make(map[uint]int)
	for index, problemID := range problems {
		problemIDMap[problemID] = index
	}

	type balloon struct {
		ID 		int 	`json:"id"`
		UserID 	uint 	`json:"user_id"`
		Nick 	string 	`json:"nick"`
		ProblemID 	int 	`json:"problem_id"`
		Color 	string 	`json:"color"`
		IsSent	bool 	`json:"is_sent"`
	}
	
	var balloons []balloon
	balloonMap := make(map[string]bool)

	for _, submit := range submits {
		newBalloon := balloon {
			UserID: submit.UserID,
			ID: submit.ID,
			Nick: submit.Nick,
			ProblemID: problemIDMap[submit.ProblemID]+1,
			Color: colors[problemIDMap[submit.ProblemID]],
		}
		submitIdentity := strconv.Itoa(int(contestIDJson.ContestID)) + strconv.Itoa(newBalloon.ProblemID)+" "+strconv.Itoa(int(newBalloon.UserID))
		if _, ok := balloonMap[submitIdentity]; ok {
			continue
		}
		balloonMap[submitIdentity] = true
		value, err := db_server.SIsNumberOfRedisSet("balloon"+strconv.Itoa(int(contestIDJson.ContestID)), submitIdentity)
		if err != nil {
			c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "Redis错误", err.Error()))
			return
		}
		newBalloon.IsSent = value
		balloons = append(balloons, newBalloon)
	}

	c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeSuccess, "获取成功", balloons))
}

func SentBalloon(c *gin.Context)  {
	if res := haveAuth(c, "setBalloonStatus"); res != common.Authed {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "权限不足", res))
		return
	}
	IDJson := struct {
		ContestID 	uint 	`json:"contest_id" form:"contest_id"`
		ProblemID 	int 	`json:"problem_id" form:"problem_id"`
		UserID 		uint 	`json:"user_id" form:"user_id"`
	}{}

	if err := c.ShouldBind(&IDJson); err != nil {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "绑定数据模型失败", err.Error()))
		return
	}

	submitIdentity := strconv.Itoa(int(IDJson.ContestID)) + strconv.Itoa(IDJson.ProblemID)+" "+strconv.Itoa(int(IDJson.UserID))

	if err := db_server.SAddToRedisSet("balloon"+strconv.Itoa(int(IDJson.ContestID)), submitIdentity); err != nil {
		c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeError, "设置失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, helper.BackendApiReturn(common.CodeSuccess, "设置成功", nil))
}