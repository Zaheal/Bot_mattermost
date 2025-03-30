package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tarantool/go-tarantool/v2"
)


func (app *application) initTarantool() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dialer := tarantool.NetDialer{
		Address: app.config.TARANTOOL_ADDRESS,
		User: app.config.TARANTOOL_USER_NAME,
		Password: app.config.TARANTOOL_USER_PASSWORD,
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	var err error
	app.mmTarantool, err = tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		app.logger.Fatal().Err(err).Msg("Connection refused")
		return err
	}

	return nil
}


func (app *application) createPoll(poll Poll) error {
	req := tarantool.NewInsertRequest("votes").Tuple([]interface{}{
		poll.ID,
		poll.Question,
		poll.Options,
		poll.Votes,
		poll.Creator,
		true,
	})

	_, err := app.mmTarantool.Do(req).Get()
	if err != nil {
		app.logger.Warn().Err(err).Msg("Can't create Poll")
		return err
	}

	return nil
}

func (app *application) getPoll(poll_id string) (Poll, error) {
	req := tarantool.NewSelectRequest("votes").
		Limit(1).
		Iterator(tarantool.IterEq).
		Key(poll_id)
	
	resp, err := app.mmTarantool.Do(req).Get()
	if err != nil {
		app.logger.Warn().Err(err).Msg("Can't get Poll")
		return Poll{}, err
	}

	if len(resp) == 0 {
		return Poll{}, fmt.Errorf("poll not found")
	}

	tuple := resp[0].([]interface{})
	poll := Poll{
		ID:       tuple[0].(string),
		Question: tuple[1].(string),
		Options:  tuple[2].([]string),
		Votes: 	  tuple[3].(map[string]int),	
		Creator:  tuple[4].(string),
		IsActive: tuple[5].(bool),
	}

	opts := tuple[2].([]interface{})
	poll.Options = make([]string, len(opts))
	for i, opt := range opts {
		poll.Options[i] = opt.(string)
	}

	votes := tuple[3].(map[interface{}]interface{})
	poll.Votes = make(map[string]int)
	for k, v := range votes {
		poll.Votes[k.(string)] = int(v.(int64))
	}

	return poll, err
}

func (app *application) updatePoll(poll_id string, option string) error {
	poll, err := app.getPoll(poll_id)
	if err != nil {
		app.logger.Warn().Err(err).Msg("Not poll for update")
		return err
	}

	if !poll.IsActive {
		return errors.New("poll is closed")
	}

	poll.Votes[option]++

	req := tarantool.NewUpdateRequest("votes").
			Key([]interface{}{poll_id}).
			Operations(tarantool.NewOperations().Assign(3, poll.Votes))
	
	resp, err := app.mmTarantool.Do(req).Get()
	if err != nil {
		app.logger.Warn().Err(err).Msg("Can't Update")
		return err
	}

	app.logger.Debug().Msg(fmt.Sprintf("Response update: %#v", resp))
	return nil
}

func (app *application) stopPoll(poll_id string, creator string) error {
	poll, err := app.getPoll(poll_id)
	if err != nil {
		app.logger.Warn().Err(err).Msg("Not poll for update")
		return err
	}

	if !poll.IsActive {
		return errors.New("poll already closed")
	}

	if poll.Creator != creator {
		return errors.New("you aren't creator this vote")
	}

	req := tarantool.NewUpdateRequest("votes").
			Key([]interface{}{poll_id}).
			Operations(tarantool.NewOperations().Assign(5, false))
	
	resp, err := app.mmTarantool.Do(req).Get()
	if err != nil {
		app.logger.Warn().Err(err).Msg("Can't Update")
		return err
	}

	app.logger.Debug().Msg(fmt.Sprintf("Response stop: %#v", resp))
	return nil
}

func (app *application) deletePoll(poll_id string) error {
	poll, err := app.getPoll(poll_id)
	if err != nil {
		app.logger.Warn().Err(err).Msg("Not poll for update")
		return err
	}

	if poll.IsActive {
		return errors.New("you can delete only inactive poll")
	}
	
	req := tarantool.NewDeleteRequest("votes").
			Key([]interface{}{poll_id})
	resp, err := app.mmTarantool.Do(req).Get()
	if err != nil {
		app.logger.Warn().Err(err).Msg("Can't Delete")
		return err
	}

	app.logger.Debug().Msg(fmt.Sprintf("Response delete: %#v", resp))
	return nil
}