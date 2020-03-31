package slackduty

import (
	"sync"
	"testing"
)

func TestAdd(t *testing.T) {
	tcs := map[string]struct {
		members    []Member
		duplicated int
	}{
		"single member":             {[]Member{{"id1", "id1@example.com"}}, 0},
		"mutiple unique member":     {[]Member{{"id1", "id1@example.com"}, {"id2", "id2@example.com"}}, 0},
		"mutiple deplicated member": {[]Member{{"id1", "id1@example.com"}, {"id1", "id1@example.com"}}, 1},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			members := &Members{}
			wg := &sync.WaitGroup{}
			for _, m := range tc.members {
				wg.Add(1)
				m := m
				go func() {
					members.Add(&m)
					wg.Done()
				}()
			}

			wg.Wait()
			if want := (len(tc.members) - tc.duplicated); len(members.Members) != want {
				t.Fatalf("members dosen't match got: %d want: %d", len(members.Members), want)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tcs := map[string]struct {
		members   []Member
		blacklist []string
		want      []Member
		success   bool
	}{
		"no blacklist":           {[]Member{{"id1", "id1@example.com"}}, []string{}, []Member{{"id1", "id1@example.com"}}, true},
		"single ID blacklist":    {[]Member{{"id1", "id1@example.com"}}, []string{"id:id1"}, []Member{}, true},
		"single Email blacklist": {[]Member{{"id1", "id1@example.com"}}, []string{"email:id1@example.com"}, []Member{}, true},
		"no matching blacklist":  {[]Member{{"id1", "id1@example.com"}}, []string{"id:idX"}, []Member{{"id1", "id1@example.com"}}, true},
		"wrong blacklist format": {[]Member{{"id1", "id1@example.com"}}, []string{"wrong:wrong"}, []Member{{"id1", "id1@example.com"}}, false},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			members := &Members{Members: tc.members}
			r, err := members.Filter(tc.blacklist)
			if err != nil {
				if tc.success {
					t.Fatalf("error: %v", err)
				} else {
					return
				}
			}

			if len(r.Members) != len(tc.want) {
				t.Fatalf("filter result is unexpected got: %d, want: %d", len(r.Members), len(tc.want))
			}
		})
	}
}

func TestFlattenMembers(t *testing.T) {
	tcs := map[string]struct {
		members []Member
		want    string
	}{
		"single member":    {[]Member{{"id1", "id1@example.com"}}, "id1"},
		"multiple members": {[]Member{{"id1", "id1@example.com"}, {"id2", "id2@example.com"}, {"id3", "id3@example.com"}}, "id1,id2,id3"},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := FlattenMembers(tc.members)
			if r != tc.want {
				t.Fatalf("result doesn't match  got: %s want: %s", r, tc.want)
			}
		})
	}
}
