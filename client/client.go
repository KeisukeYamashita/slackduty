package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/KeisukeYamashita/slackduty/config"
	"github.com/KeisukeYamashita/slackduty/slackduty"
	"github.com/robfig/cron"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Client acts likes a manager of the jobs.
type Client struct {
	config          *config.Config
	externalTrigger bool
	cron            *cron.Cron
	pagerduty       PagerdutyClient
	slack           SlackClient
	logger          *zap.Logger
}

type options struct {
	externalTrigger bool
}

// ClientOption are options that developers can configure for the
// Slackduty clients
type ClientOption func(*options)

// WithExternalTrigger should be used when this process is kicked by
// external resources(e.g. Cloud Scheduler, Unix crontab).
func WithExternalTrigger() ClientOption {
	return func(o *options) {
		o.externalTrigger = true
	}
}

// New creates a Client for Slack & PagerDuty API.
func New(cfg *config.Config, pdAPIKey, slackAPIKey string, logger *zap.Logger, opts ...ClientOption) *Client {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	pdClient := NewPagerDutyClient(pdAPIKey)
	slackClient := NewSlackClient(slackAPIKey)
	c := &Client{
		config:    cfg,
		pagerduty: pdClient,
		slack:     slackClient,
		logger:    logger,
	}

	if o.externalTrigger {
		c.externalTrigger = o.externalTrigger
	} else {
		c.cron = cron.New()
	}

	return c
}

// Run the job.
func (c *Client) Run() error {
	defer func() {
		if !c.externalTrigger {
			c.logger.Info("stop cronjob")
			c.cron.Stop()
		}
	}()

	wg := &sync.WaitGroup{}
	for _, group := range c.config.Groups {
		wg.Add(1)
		group := group
		go func() {
			opts := []slackduty.JobOption{}
			if !c.externalTrigger {
				opts = append(opts, slackduty.WithSchedule(c.cron, group.Schedule))
			}

			job := slackduty.NewJob(func() error { return c.configureGroup(&group) }, opts...)
			ctx := context.Background()
			err := job.Run(ctx)
			if err != nil {
				c.logger.Error("failed to update Slack usergroup", zap.Error(err), zap.Bool("external trigger", c.externalTrigger))
			}
			c.logger.Info("successfully ran a job for updating Slack usergroup", zap.String("group", group.Name), zap.Bool("external trigger", c.externalTrigger))
			wg.Done()
		}()
	}

	if !c.externalTrigger {
		c.logger.Info("start cronjob", zap.Int("group count", len(c.config.Groups)), zap.Bool("external trigger", c.externalTrigger))
		c.cron.Start()
	} else {
		c.logger.Info("start job", zap.Int("group count", len(c.config.Groups)))
	}

	wg.Wait()
	return nil
}

func (c *Client) configureGroup(group *config.Group) error {
	c.logger.Info("start to run configure group job", zap.String("name", group.Name), zap.String("schedule", group.Schedule))

	if err := c.preCheck(group); err != nil {
		c.logger.Error("precheck failed", zap.Error(err), zap.String("group", group.Name), zap.String("schedule", group.Schedule))
		return fmt.Errorf("precheck failed error: %v", err)
	}

	members, err := c.GetMembers(group.Members)
	if err != nil {
		c.logger.Error("failed to get members of the group", zap.Error(err), zap.String("group", group.Name), zap.String("schedule", group.Schedule))
		return err
	}

	members, err = members.Filter(group.Exclude)
	if err != nil {
		c.logger.Error("failed to filter the members of the group", zap.Error(err), zap.String("group", group.Name), zap.String("schedule", group.Schedule))
		return err
	}

	if len(members.Members) == 0 {
		c.logger.Warn("no member was in the member", zap.String("group", group.Name), zap.String("schedule", group.Schedule))
		return nil
	}

	for _, usergroup := range group.Usergroups {
		if err := c.updateUsergroup(usergroup, members.Members); err != nil {
			c.logger.Error("failed to update the Slack usergroup", zap.Error(err), zap.String("group", group.Name))
			return err
		}

		c.logger.Info("updated a slack usergroup", zap.String("group", group.Name), zap.String("schedule", group.Schedule), zap.String("usergroup", usergroup))
	}

	return nil
}

// GetMembers get all members that should be a member of the usergroup(s)
// For can specify Slack user and Pagerduty users, teams, services and also
// schedules.
func (c *Client) GetMembers(cfg *config.Members) (*slackduty.Members, error) {
	members := &slackduty.Members{}
	eg := errgroup.Group{}

	if cfg.Pagerduty != nil {
		eg.Go(func() error {
			err := c.getPagerDutyMembers(cfg.Pagerduty, members)
			if err != nil {
				c.logger.Error("failed to get PagerDuty members", zap.Error(err))
				return err
			}

			return nil
		})
	}

	if cfg.Slack != nil {
		eg.Go(func() error {
			err := c.getSlackUsers(cfg.Slack, members)
			if err != nil {
				c.logger.Error("failed to get Slack members", zap.Error(err))
				return err
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return members, nil
}

func (c *Client) preCheck(group *config.Group) error {
	c.logger.Info("precheck started", zap.String("group", group.Name), zap.String("schedule", group.Schedule))
	ugs, err := c.slack.GetUsergroups()
	if err != nil {
		c.logger.Info("precheck failed to get Slack usergroups", zap.Error(err))
		return err
	}

	var exists bool

LOOP:
	for _, ug := range ugs {
		for _, usergroup := range group.Usergroups {
			s := strings.Split(usergroup, ":")
			if len(s) != 2 {
				return fmt.Errorf("usergroups is specified in wrong format user: %s", usergroup)
			}

			kind := s[0]
			val := s[1]

			switch kind {
			case "handle":
				if ug.Handle == val {
					exists = true
					break LOOP
				}
			}
		}
	}

	if !exists {
		return errors.New("slack user group handle doesn't exists")
	}

	c.logger.Info("precheck success", zap.String("group", group.Name), zap.String("schedule", group.Schedule))
	return nil
}

func (c *Client) getPagerDutyMembers(pdConfig *config.Pagerduty, members *slackduty.Members) error {
	eg := errgroup.Group{}
	eg.Go(func() error {
		err := c.getPagerdutySchedules(pdConfig.Schedules, members)
		if err != nil {
			c.logger.Error("failed to get PagerDuty schedules members", zap.Error(err))
			return err
		}

		return nil
	})

	eg.Go(func() error {
		err := c.getPagerdutyServices(pdConfig.Services, members)
		if err != nil {
			c.logger.Error("failed to get PagerDuty services members", zap.Error(err))
			return err
		}

		return nil
	})

	eg.Go(func() error {
		err := c.getPagerdutyTeams(pdConfig.Teams, members)
		if err != nil {
			c.logger.Error("failed to get PagerDuty teams members", zap.Error(err))
			return err
		}

		return nil
	})

	eg.Go(func() error {
		err := c.getPagerdutyUsers(pdConfig.Users, members)
		c.logger.Error("failed to get PagerDuty users members", zap.Error(err))
		if err != nil {
			return err
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *Client) getPagerdutySchedules(schedules []string, members *slackduty.Members) error {
	eg := errgroup.Group{}
	for _, schedule := range schedules {
		schedule := schedule
		eg.Go(func() error {
			pdUsers, err := c.pagerduty.GetScheduledUser(schedule)
			if err != nil {
				return err
			}

			for _, pdUser := range pdUsers {
				slackUser, err := c.slack.GetUser(fmt.Sprintf("email:%s", pdUser.Email))
				if err != nil {
					return err
				}

				member := convSlackUser(slackUser, pdUser.Email)
				members.Add(member)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *Client) getPagerdutyServices(svcs []string, members *slackduty.Members) error {
	eg := errgroup.Group{}
	for _, svc := range svcs {
		svc := svc
		eg.Go(func() error {
			pdUsers, err := c.pagerduty.GetService(svc)
			if err != nil {
				return err
			}

			for _, pdUser := range pdUsers {
				slackUser, err := c.slack.GetUser(fmt.Sprintf("email:%s", pdUser.Email))
				if err != nil {
					return err
				}

				member := convSlackUser(slackUser, pdUser.Email)
				members.Add(member)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *Client) getPagerdutyTeams(teams []string, members *slackduty.Members) error {
	eg := errgroup.Group{}
	for _, team := range teams {
		team := team
		eg.Go(func() error {
			pdUsers, err := c.pagerduty.GetTeam(team)
			if err != nil {
				return err
			}

			for _, pdUser := range pdUsers {
				slackUser, err := c.slack.GetUser(fmt.Sprintf("email:%s", pdUser.Email))
				if err != nil {
					return err
				}

				member := convSlackUser(slackUser, pdUser.Email)
				members.Add(member)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *Client) getPagerdutyUsers(users []string, members *slackduty.Members) error {
	eg := errgroup.Group{}
	for _, user := range users {
		user := user
		eg.Go(func() error {
			pdUser, err := c.pagerduty.GetUser(user)
			if err != nil {
				return err
			}

			slackUser, err := c.slack.GetUser(fmt.Sprintf("email:%s", pdUser.Email))
			if err != nil {
				return err
			}

			member := convSlackUser(slackUser, pdUser.Email)
			members.Add(member)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *Client) getSlackUsers(users *config.Slack, members *slackduty.Members) error {
	eg := errgroup.Group{}
	for _, user := range *users {
		user := user
		eg.Go(func() error {
			slackUser, err := c.slack.GetUser(user)
			if err != nil {
				return err
			}

			var email string
			if s := strings.Split(user, ":"); s[0] == "email" {
				email = s[1]
			}

			member := convSlackUser(slackUser, email)
			members.Add(member)

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *Client) updateUsergroup(handle string, members []slackduty.Member) error {
	flatMembers := slackduty.FlattenMembers(members)
	err := c.slack.UpdateUsergroup(handle, flatMembers)
	return err
}
