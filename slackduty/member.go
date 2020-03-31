package slackduty

import (
	"fmt"
	"strings"
	"sync"
)

// Members is a struct for managing the Members from
// various PagerDuty resources(e.g. teams, services, schedules).
type Members struct {
	mux     sync.RWMutex
	Members []Member
}

// Member represents a single member(Slack user).
// Currently, we use ID and Email only.
type Member struct {
	ID    string
	Email string
}

// Add appends a member to the Members struct.
// It will also removes the duplication of the Slack ID.
func (m *Members) Add(member *Member) {
	m.mux.Lock()
	var exists bool
	for _, m := range m.Members {
		if m.ID == member.ID {
			exists = true
		}
	}

	if !exists {
		m.Members = append(m.Members, *member)
	}
	m.mux.Unlock()
}

// Filter removes the excluded Slack users by ID or Email.
func (m *Members) Filter(blacklists []string) (*Members, error) {
	newMembers := &Members{}
	if len(blacklists) == 0 {
		newMembers.Members = m.Members
		return newMembers, nil
	}

	for _, member := range m.Members {
		for _, blacklist := range blacklists {
			s := strings.Split(blacklist, ":")
			if len(s) != 2 {
				return nil, fmt.Errorf("exclude is specified in wrong format: %s", blacklist)
			}

			kind := s[0]
			val := s[1]

			switch kind {
			case "id":
				if member.ID != val {
					newMembers.Members = append(newMembers.Members, member)
				}
			case "email":
				if member.Email != val {
					newMembers.Members = append(newMembers.Members, member)
				}
			default:
				return nil, fmt.Errorf("blacklist format should be id or email blacklist: %s", blacklist)
			}
		}
	}

	return newMembers, nil
}

// FlattenMembers returns a single string with ID splited by members ID.
// It is intended to use for Slack API usergroup.user.update to update the
// members of the Slack usergroup.
func FlattenMembers(members []Member) string {
	result := ""
	for i, member := range members {
		if i != 0 {
			result += fmt.Sprintf(",%s", member.ID)
			continue
		} else {
			result = member.ID
		}
	}

	return result
}
