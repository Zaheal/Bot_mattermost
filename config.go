package main

import (
	"os"
)


type Config struct {
	ACCESS_TOKEN   			string
	MATTERMOST_URL 			string
	TEAM_NAME 	   			string
	CHANNEL_NAME   			string
	CHANNEL_ID     			string
	TARANTOOL_USER_NAME 	string
	TARANTOOL_USER_PASSWORD string
	TARANTOOL_ADDRESS       string
}


func getConfig() Config {
	var settings Config
	
	settings.ACCESS_TOKEN   		 = getEnv("ACCESS_TOKEN", "")
	settings.MATTERMOST_URL 		 = getEnv("MATTERMOST_URL", "")
	settings.TEAM_NAME      		 = getEnv("TEAM_NAME", "")
	settings.CHANNEL_NAME   		 = getEnv("CHANNEL_NAME", "")
	settings.CHANNEL_ID     		 = getEnv("CHANNEL_ID", "")
	settings.TARANTOOL_USER_NAME 	 = getEnv("TARANTOOL_USER_NAME", "")
	settings.TARANTOOL_USER_PASSWORD = getEnv("TARANTOOL_USER_PASSWORD", "")
	settings.TARANTOOL_ADDRESS       = getEnv("TARANTOOL_ADDRESS", "")

	return settings
}


func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
