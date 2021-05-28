package controller

import (
	"OnlineJudge/app/api/model"
	"OnlineJudge/app/common"
	"OnlineJudge/app/common/validate"
	"OnlineJudge/app/helper"
	"OnlineJudge/config"
	"OnlineJudge/db_server"
	"OnlineJudge/judger"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func Submit(c *gin.Context) {

	res := checkLogin(c)
	if res.Status == common.CodeError {
		c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
		return
	}

	submitModel := model.Submit{}
	submitValidate := validate.SubmitValidate
	session := sessions.Default(c)
	userID := session.Get("user_id")

	if userID == nil {
		c.JSON(http.StatusOK, helper.ApiReturn(common.CodeError, "用户未登录", ""))
		return
	}

	format := "2006-01-02 15:04:05"
	now, _ := time.Parse(format, time.Now().Format(format))
	interval := config.GetWutOjConfig()["interval_time"].(int)
	redisStr := "user_last_submit" + strconv.Itoa(int(userID.(uint)))
	if value, err := db_server.GetFromRedis(redisStr); err == nil {
		defaultFormat := "2006-01-02 15:04:05 +0000 UTC"
		lastStr, _ := redis.String(value, err)
		last, _ := time.Parse(defaultFormat, lastStr)
		fmt.Printf("now: %v, last: %v\n", now, last)

		if now.Unix()-last.Unix() <= int64(interval) {
			c.JSON(http.StatusOK, helper.ApiReturn(common.CodeError, "交题间隔过快，请五秒后再试", ""))
			return
		}
	}
	_ = db_server.PutToRedis(redisStr, now, 3600)

	var submitJson struct {
		Language   string `json:"language"`
		SourceCode string `json:"source_code"`
		ProblemID  uint   `json:"problem_id"`
		ContestID  uint   `json:"contest_id"`
	}

	if err := c.ShouldBindJSON(&submitJson); err != nil {
		c.JSON(http.StatusOK, helper.ApiReturn(common.CodeError, "数据绑定模型错误", err.Error()))
		return
	}

	submitMap := helper.Struct2Map(submitJson)

	if res, err := submitValidate.ValidateMap(submitMap, "add"); !res {
		c.JSON(http.StatusOK, helper.ApiReturn(common.CodeError, err.Error(), ""))
		return
	}

	if helper.LanguageID(submitJson.Language) == -1 {
		c.JSON(http.StatusOK, helper.ApiReturn(common.CodeError, "不支持的语言类型", nil))
		return
	}

	newSubmit := model.Submit{
		UserID:     userID.(uint),
		Nick:       GetUserNickFromSession(c),
		Language:   helper.LanguageID(submitJson.Language),
		SourceCode: submitJson.SourceCode,
		ProblemID:  submitJson.ProblemID,
		ContestID:  submitJson.ContestID,
		Status:     "Judging",
		SubmitTime: now,
	}

	resp := submitModel.AddSubmit(&newSubmit)

	go func(submit model.Submit) {
		judge(submit)
	}(newSubmit)

	c.JSON(http.StatusOK, helper.ApiReturn(resp.Status, resp.Msg, resp.Data))
	return

}

func judge(submit model.Submit) {
	langConfig := config.GetLangConfigs()[submit.Language]
	submitData := judger.SubmitData{
		Id:           uint64(submit.ID),
		Pid:          uint64(submit.ProblemID),
		Language:     langConfig.Lang,
		Code:         submit.SourceCode,
		BuildScript:  langConfig.BuildSh,
		RunnerConfig: langConfig.RunnerConfig,
	}

	callback := func(id uint64, result judger.JudgeResult) {
		// Put Result To DB
		// TODO if set contest, update redis rank info
		if result.IsFinished {
			data := map[string]interface{}{
				"status": result.Status,
				"time":   result.Time,
				"memory": result.Memory,
				"msg":    result.Msg,
			}
			submitModel := model.Submit{}
			submitModel.UpdateStatusAfterSubmit(int(id), data)
			if submit.ContestID != 0 {
				// set to redis
				beginTime, _, frozenTime, err := getContestTime(submit.ContestID)
				if err != nil {
					return
				}
				format := "2006-01-02 15:04:05"
				now, _ := time.Parse(format, time.Now().Format(format))
				fmt.Printf("Submit Time: %v, BeginTime: %v, FrozenTime: %v\n", now, beginTime, frozenTime)
				if now.Unix() < beginTime.Unix() || now.Unix() > frozenTime.Unix() {
					return
				}
				user := user{UserID: submit.UserID, Nick: submit.Nick, Penalty: 0, ACNum: 0, ProblemID: make(map[uint]problem)}
				if itemStr, err := redis.String(db_server.GetFromRedis("contest_rank" + strconv.Itoa(int(submit.ContestID)) + "user_id" + strconv.Itoa(int(submit.UserID)))); err == nil {
					_ = json.Unmarshal([]byte(itemStr), &user)
				}
				if _, ok := user.ProblemID[submit.ProblemID]; !ok {
					user.ProblemID[submit.ProblemID] = problem{SuccessTime: 0, Times: 0}
				}
				userProblem := user.ProblemID[submit.ProblemID]
				log.Println(result)
				if user.ProblemID[submit.ProblemID].SuccessTime == 0 {
					if result.Status == "AC" {
						user.ProblemID[submit.ProblemID] = problem{SuccessTime: submit.SubmitTime.Unix() - beginTime.Unix(), Times: userProblem.Times + 1}
						user.ACNum++
						user.Penalty += int64(userProblem.Times*20*60) + user.ProblemID[submit.ProblemID].SuccessTime
					} else if result.Status != "CE" && result.Status != "UE" {
						user.ProblemID[submit.ProblemID] = problem{SuccessTime: 0, Times: userProblem.Times + 1}
					}

					itemStr, _ := json.Marshal(user)
					_ = db_server.PutToRedis("contest_rank"+strconv.Itoa(int(submit.ContestID))+"user_id"+strconv.Itoa(int(user.UserID)), itemStr, 3600)
					score := -int64(user.ACNum)*1000000000 + user.Penalty
					_ = db_server.ZAddToRedis("contest_rank"+strconv.Itoa(int(submit.ContestID)), score, user.UserID)

				}
			}
		}
	}

	instance := judger.GetInstance()

	go instance.Submit(submitData, callback)
}

func GetSubmitInfo(c *gin.Context) {

	res := checkLogin(c)
	if res.Status == common.CodeError {
		c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
		return
	}

	submitModel := model.Submit{}

	submitJson := struct {
		PageNumber int `json:"page_number" form:"page_number"`
	}{}

	ConfigMap := config.GetWutOjConfig()
	Limit := ConfigMap["page_limit"].(int)

	UserID := GetUserIdFromSession(c)

	if c.ShouldBind(&submitJson) == nil {
		res := submitModel.GetAllSubmit(Limit*(submitJson.PageNumber-1), Limit, UserID)
		c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
		return
	}

	c.JSON(http.StatusOK, helper.ApiReturn(common.CodeError, "绑定数据模型失败", false))

}

// TODO
func GetAllSubmitInfo(c *gin.Context) {
	// change to GetSubmitInfo
}

// TODO
func GetProblemSubmitInfo(c *gin.Context) {
	res := checkLogin(c)
	if res.Status == common.CodeError {
		c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
		return
	}

	submitModel := model.Submit{}
	submitJson := model.Submit{}

	if c.ShouldBindQuery(&submitJson) != nil {
		c.JSON(http.StatusOK, helper.ApiReturn(common.CodeError, "绑定数据模型失败", false))
		return
	}

	submitJson.UserID = GetUserIdFromSession(c)
	res = submitModel.GetProblemSubmit(submitJson)
	c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
	return
}

// TODO
func GetUserContestSubmitInfo(c *gin.Context) {
	res := checkLogin(c)
	if res.Status == common.CodeError {
		c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
		return
	}

	submitValidate := validate.SubmitValidate
	submitModel := model.Submit{}

	submitJson := struct {
		ContestID  uint `json:"contest_id" form:"contest_id" uri:"contest_id"`
		PageNumber int  `json:"page_number" form:"page_number" uri:"page_number"`
		UserID     uint
	}{}

	if c.ShouldBindQuery(&submitJson) != nil {
		c.JSON(http.StatusOK, helper.ApiReturn(common.CodeError, "绑定数据模型失败", false))
		return
	}

	submitJson.UserID = GetUserIdFromSession(c)

	submitMap := helper.Struct2Map(submitJson)
	if res, err := submitValidate.ValidateMap(submitMap, "contest_log"); !res {
		c.JSON(http.StatusOK, helper.ApiReturn(common.CodeError, err.Error(), 0))
		return
	}

	res = submitModel.GetContestSubmit(submitJson.UserID, submitJson.ContestID, submitJson.PageNumber)
	c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
	return
}

func GetSubmitByID(c *gin.Context) {

	res := checkLogin(c)
	if res.Status == common.CodeError {
		c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
		return
	}

	submitJson := model.Submit{}
	submitModel := model.Submit{}
	submitValidate := validate.SubmitValidate

	if c.ShouldBindQuery(&submitJson) != nil {
		c.JSON(http.StatusOK, helper.ApiReturn(common.CodeError, "绑定数据模型失败", false))
		return
	}

	submitMap := helper.Struct2Map(submitJson)
	if res, err := submitValidate.ValidateMap(submitMap, "find"); !res {
		c.JSON(http.StatusOK, helper.ApiReturn(common.CodeError, err.Error(), 0))
		return
	}

	submitJson.UserID = GetUserIdFromSession(c)

	res = submitModel.GetSubmitByID(submitJson.ID, submitJson.UserID)
	c.JSON(http.StatusOK, helper.ApiReturn(res.Status, res.Msg, res.Data))
	return
}
