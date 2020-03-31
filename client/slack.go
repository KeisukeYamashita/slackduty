package client

import (
	"fmt"
	"strings"

	"github.com/KeisukeYamashita/slackduty/slackduty"
	"github.com/slack-go/slack"
)

// SlackClient is a interface that the Slack client should implement
type SlackClient interface {
	CreateUsergroup() error
	GetUser(string) (*slack.User, error)
	GetUsergroups() ([]slack.UserGroup, error)
	UpdateUsergroup(string, string) error
}

var _ SlackClient = (*slackClient)(nil)

type slackClient struct {
	client *slack.Client
}

// NewSlackClient creates a new Slack API client
func NewSlackClient(apiKey string) SlackClient {
	client := slack.New(apiKey)

	return &slackClient{
		client: client,
	}
}

func (c *slackClient) CreateUsergroup() error {
	return nil
}

func (c *slackClient) GetUser(user string) (*slack.User, error) {
	s := strings.Split(user, ":")
	if len(s) != 2 {
		return nil, fmt.Errorf("user is specified in wrong format user: %s", user)
	}

	kind := s[0]
	val := s[1]

	switch kind {
	case "id":
		// Note(KeisukeYamashita): To update Slack usergroup, the user ID is only required.
		// Therefore, we don't need to access the Slack API for future info and I early return with the slack.User struct
		return &slack.User{ID: val}, nil
	case "email":
		slackUser, err := c.client.GetUserByEmail(val)
		if err != nil {
			return nil, err
		}

		return slackUser, nil
	default:
		return nil, fmt.Errorf("user kind %s is invalid, must be email, id for user: %s", kind, user)
	}
}

func (c *slackClient) GetUsergroups() ([]slack.UserGroup, error) {
	return c.client.GetUserGroups()
}

func (c *slackClient) UpdateUsergroup(handle string, members string) error {
	s := strings.Split(handle, ":")
	if len(s) != 2 {
		return fmt.Errorf("handle is specified in wrong format handle: %s", handle)
	}

	kind := s[0]
	val := s[1]

	var groupID string
	switch kind {
	case "id":
		groupID = val
	case "handle":
		ugs, err := c.client.GetUserGroups()
		if err != nil {
			return err
		}

		var exists bool
		var usergroup *slack.UserGroup

	LOOP:
		for _, ug := range ugs {
			if ug.Handle == val {
				exists = true
				usergroup = &ug
				break LOOP
			}
		}

		if !exists {
			return fmt.Errorf("usergroup doesn't exists for handle: %s", handle)
		}

		groupID = usergroup.ID
	default:
		return fmt.Errorf("handle kind %s is invalid, must be id for name: %s", kind, handle)
	}

	_, err := c.client.UpdateUserGroupMembers(groupID, members)
	return err
}

func convSlackUser(user *slack.User, email string) *slackduty.Member {
	return &slackduty.Member{
		ID:    user.ID,
		Email: email,
	}
}
