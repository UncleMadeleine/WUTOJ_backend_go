package validate

import (
	"OnlineJudge/app/helper"
)

var UserValidate helper.Validator

func init() {
	rules := map[string]string{
		"user_id":        "required",
		"nick":           "required|maxLen:25",
		"password":       "required|minLen:6|maxLen:16",
		"old_password":   "required|minLen:6|maxLen:16",
		"password_check": "required|minLen:6|maxLen:16",
		"identify":       "int:-1,3",
		"realname":       "required",
		"school":         "required",
		"major":          "required",
		"class":          "required",
		"contact":        "required",
		"mail":           "required|email",
		"check":          "required",
		"status":         "required|int:-1,0",
		"is_admin":       "required|bool",
		"verify_code":    "required",
		"param":          "required",
		"users": 			"required",
	}

	scenes := map[string][]string{
		"addUser":         {"nick", "password", "realname", "school", "major", "class", "contact", "mail", "status"},
		"editUser":        {"nick", "realname", "school", "major", "class", "contact", "mail", "status"},
		"searchUser_id":   {"user_id"},
		"searchUser_nick": {"nick"},
		"search":          {"param"},
		"deleteUser":      {"user_id"},
		"updateUser": 		{"user_id"},
		"groupDelete": 		{"users"},
		"foreAddUser":     {},
		"login":           {"nick", "password"},
		"register":        {"nick", "password", "password_check", "realname", "school", "major", "class", "contact", "mail"},
		"forget":          {"mail"},
		"forget_password": {"nick", "password", "password_check", "check"},
		"change_password": {"nick", "old_password", "password", "password_check"},
		"set_admin":       {"uid", "is_admin"},
		"check_login":     {"user_id"},
		"update_password": {"mail", "verify_code", "password", "password_check"},
		"update_info":     {},
	}
	UserValidate.Rules = rules
	UserValidate.Scenes = scenes
}
