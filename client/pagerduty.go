package client

import (
	"fmt"
	"strings"
	"sync"

	"github.com/PagerDuty/go-pagerduty"
	"golang.org/x/sync/errgroup"
)

// PagerdutyClient is a interface that the PagerDuty client should implement
type PagerdutyClient interface {
	GetScheduledUser(string) ([]pagerduty.User, error)
	GetService(string) ([]pagerduty.User, error)
	GetTeam(string) ([]pagerduty.User, error)
	GetUser(string) (*pagerduty.User, error)
}

var _ PagerdutyClient = (*pagerdutyClient)(nil)

type pagerdutyClient struct {
	client *pagerduty.Client
}

// NewPagerDutyClient creates a new PagerDuty API client
func NewPagerDutyClient(apiKey string) PagerdutyClient {
	client := pagerduty.NewClient(apiKey)
	return &pagerdutyClient{
		client: client,
	}
}

func (c *pagerdutyClient) GetScheduledUser(schedule string) ([]pagerduty.User, error) {
	s := strings.Split(schedule, ":")
	if len(s) != 2 {
		return nil, fmt.Errorf("schedule is specified in wrong format schedule: %s", schedule)
	}

	kind := s[0]
	val := s[1]

	var pdSche *pagerduty.Schedule
	var err error
	switch kind {
	case "id":
		opt := pagerduty.GetScheduleOptions{}
		pdSche, err = c.client.GetSchedule(val, opt)
	case "name":
		opt := pagerduty.ListSchedulesOptions{Query: val}
		resp, err := c.client.ListSchedules(opt)
		if err != nil {
			return nil, err
		}

		pdSches := resp.Schedules
		if c := len(pdSches); c != 1 {
			if c == 0 {
				return nil, fmt.Errorf("no schedule exists for schedule: %s", schedule)
			}

			if c > 1 {
				return nil, fmt.Errorf("more than one schedules exists for schedule name %s got %d schedules, team: %s", val, len(pdSches), schedule)
			}
		}

		pdSche = &pdSches[0]
	default:
		return nil, fmt.Errorf("schedule kind %s is invalid, must be email, id or name for schedule:%s", kind, schedule)
	}

	if err != nil {
		return nil, err
	}

	users := []pagerduty.User{}
	for _, apiObj := range pdSche.Users {
		opt := pagerduty.GetUserOptions{}
		user, err := c.client.GetUser(apiObj.ID, opt)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}

	return users, nil
}

func (c *pagerdutyClient) GetService(service string) ([]pagerduty.User, error) {
	s := strings.Split(service, ":")
	if len(s) != 2 {
		return nil, fmt.Errorf("service is specified in wrong format service: %s", service)
	}

	kind := s[0]
	val := s[1]

	var pdSvc *pagerduty.Service
	var err error
	switch kind {
	case "id":
		opt := pagerduty.GetServiceOptions{}
		pdSvc, err = c.client.GetService(val, &opt)
	case "name":
		opt := pagerduty.ListServiceOptions{Query: val}
		resp, err := c.client.ListServices(opt)
		if err != nil {
			return nil, err
		}

		pdSvcs := resp.Services
		if c := len(pdSvcs); c != 1 {
			if c == 0 {
				return nil, fmt.Errorf("no service exists for service: %s", service)
			}

			if c > 1 {
				return nil, fmt.Errorf("more than one services exists for service name %s got %d services, team: %s", val, len(pdSvcs), service)
			}
		}

		pdSvc = &pdSvcs[0]
	default:
		return nil, fmt.Errorf("service kind %s is invalid, must be email, id or name for service:%s", kind, service)
	}

	if err != nil {
		return nil, err
	}

	eg := errgroup.Group{}
	var mux sync.Mutex
	users := []pagerduty.User{}
	for _, team := range pdSvc.Teams {
		team := team
		eg.Go(func() error {
			svcUsers, err := c.GetTeam(fmt.Sprintf("id:%s", team.ID))
			if err != nil {
				return err
			}

			mux.Lock()
			users = append(users, svcUsers...)
			mux.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *pagerdutyClient) GetTeam(team string) ([]pagerduty.User, error) {
	s := strings.Split(team, ":")
	if len(s) != 2 {
		return nil, fmt.Errorf("team is specified in wrong format team: %s", team)
	}

	kind := s[0]
	val := s[1]

	var members []pagerduty.Member
	var err error
	switch kind {
	case "id":
		members, err = c.client.ListAllMembers(val)
	case "name":
		opt := pagerduty.ListTeamOptions{Query: val}
		resp, err := c.client.ListTeams(opt)
		if err != nil {
			return nil, err
		}

		teams := resp.Teams
		if c := len(teams); c != 1 {
			if c == 0 {
				return nil, fmt.Errorf("no team exists team: %s", team)
			}

			if c > 1 {
				return nil, fmt.Errorf("more than one team exists for team name %s, team: %s", val, team)
			}
		}

		id := &teams[0].ID
		members, err = c.client.ListAllMembers(*id)
	default:
		return nil, fmt.Errorf("team kind %s is invalid, must be email, id or name for team:%s", kind, team)
	}

	if err != nil {
		return nil, err
	}

	users := []pagerduty.User{}
	for _, member := range members {
		opt := pagerduty.GetUserOptions{}
		user, err := c.client.GetUser(member.APIObject.ID, opt)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}

	return users, nil
}

func (c *pagerdutyClient) GetUser(user string) (*pagerduty.User, error) {
	s := strings.Split(user, ":")
	if len(s) != 2 {
		return nil, fmt.Errorf("user is specified in wrong format user: %s", user)
	}

	kind := s[0]
	val := s[1]

	switch kind {
	case "id":
		opt := pagerduty.GetUserOptions{Includes: []string{"contact_methods"}}
		return c.client.GetUser(val, opt)
	case "name", "email":
		opt := pagerduty.ListUsersOptions{Query: val, Includes: []string{"contact_methods"}}
		resp, err := c.client.ListUsers(opt)
		if err != nil {
			return nil, err
		}

		users := resp.Users
		if c := len(users); c != 1 {
			if c == 0 {
				return nil, fmt.Errorf("no user exists user: %s", user)
			}

			if c > 1 {
				return nil, fmt.Errorf("more than one user exists for user name %s got %d users, user: %s", val, len(users), user)
			}
		}

		return &users[0], nil
	default:
		return nil, fmt.Errorf("user kind %s is invalid, must be email, id or name for user:%s", kind, user)
	}
}
