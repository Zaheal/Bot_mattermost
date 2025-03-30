package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
)

type Poll struct {
	ID 		  string
	Question  string
	Options   []string
	Votes 	  map[string]int
	Creator   string
	IsActive  bool
}

func (app *application) handlerVoteCommands(post *model.Post) {
	msg := strings.TrimSpace(post.Message)
	parts := strings.Fields(msg)

	if len(parts) < 2 {
		app.sendHelp(post.Id)
		return
	}


	command := parts[1]

	switch command {
	case "create":
		err := app.handlerCreateVote(post.Id, parts[2], parts[3:], post.UserId)
		if err != nil {
			sendMsg(app, "Произошла ошибка", post.ChannelId)
		} else {
			sendMsg(app, fmt.Sprintf(
				"Создан опрос:\nId: %s\nВопрос: %s\nОтветы:\n- %s",
				post.Id,
				parts[2],
				strings.Join(parts[3:], "\n-"),
			), post.ChannelId)
		}
	case "vote":
		err := app.handlerVote(parts[2], parts[3])
		if err != nil {
			sendMsg(app, err.Error(), post.ChannelId)
		} else {
			sendMsg(app, "Your vote is very important to us", post.ChannelId)
		}
	case "results":
		data, err := app.handlerResultsVote(parts[2])
		if err != nil {
			sendMsg(app, "Произошла ошибка", post.ChannelId)
		} else {
			var builder strings.Builder
			builder.WriteString(fmt.Sprintf("%s\n", data.Question))

			for key, value := range data.Votes {
				builder.WriteString(fmt.Sprintf("- %s: %b", key, value))
			}
			sendMsg(app, builder.String(), post.ChannelId)
		}
	case "end":
		err := app.handlerStopVote(parts[2], post.UserId)
		if err != nil {
			sendMsg(app, err.Error(), post.ChannelId)
		} else {
			sendMsg(app, "Poll stoped", post.ChannelId)
		}
	case "delete":
		err := app.handlerDeleteVote(parts[2])
		if err != nil {
			sendMsg(app, err.Error(), post.ChannelId)
		} else {
			sendMsg(app, "Poll deleted", post.ChannelId)
		}
	default:
		app.sendHelp(post.ChannelId)
	}
}

// Реализация создания голосования
// Формат: /vote create "Question?" "Option1" "Option2" ...
func (app *application) handlerCreateVote(poll_id string, question string, options []string, creator string) error {
	mapVotes := make(map[string]int)
	for _, item := range options {
		mapVotes[item] = 0
	}

	poll := Poll{
		ID:        poll_id,
		Question:  question,
		Options:   options,
		Votes:     mapVotes,
		Creator:   creator,
	}

	err := app.createPoll(poll)
	if err != nil {
		return err
	}

	return nil
}

// Реализация голосования
// Формат: /vote vote poll_id option
// TODO: один и тот же человек голосует сколько хочет и это не отслеживается
func (app *application) handlerVote(poll_id string, option string) error {
	err := app.updatePoll(poll_id, option)
	if err != nil {
		return err
	}

	return nil
}

// Реализация просмотра результатов
// Формат: /vote results poll_id
func (app *application) handlerResultsVote(poll_id string) (Poll, error) {
	data, err := app.getPoll(poll_id)
	if err != nil {
		return data, err
	}

	return data, err
}

// Реализация завершения голосования
// Формат: /vote end poll_id
func (app *application) handlerStopVote(poll_id string, creator string) error {
	err := app.stopPoll(poll_id, creator)
	if err != nil {
		return err
	}

	return nil
}

// Реализация удаления голосования
// Формат: /vote delete poll_id
func (app *application) handlerDeleteVote(poll_id string) error {
	err := app.deletePoll(poll_id)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) sendHelp(postID string) {
	helpText := `**Команды голосования:**
/vote create "Вопрос?" "Вариант1" "Вариант2" - Создать голосование
/vote vote <ID> <вариант> - Проголосовать
/vote results <ID> - Показать результаты
/vote end <ID> - Завершить голосование (только создатель)
/vote delete <ID> - Удалить голосование (только создатель)`
	
	sendMsg(app, helpText, postID)
}