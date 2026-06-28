package stack

import (
	"testing"
)

func makeStack(name string, tags ...string) *Stack {
	return &Stack{
		Name: name,
		Config: StackConfig{
			Name: name,
			Tags: tags,
		},
	}
}

func TestFilterByTag(t *testing.T) {
	tests := []struct {
		name   string
		stacks []*Stack
		tag    string
		want   []string // expected stack names
	}{
		{
			name:   "empty slice returns empty result",
			stacks: nil,
			tag:    "backend",
			want:   nil,
		},
		{
			name: "no match returns empty result",
			stacks: []*Stack{
				makeStack("alpha", "frontend"),
				makeStack("beta", "frontend"),
			},
			tag:  "backend",
			want: nil,
		},
		{
			name: "partial match returns only matching stacks",
			stacks: []*Stack{
				makeStack("alpha", "frontend"),
				makeStack("beta", "backend"),
				makeStack("gamma", "backend", "db"),
			},
			tag:  "backend",
			want: []string{"beta", "gamma"},
		},
		{
			name: "all match returns all stacks",
			stacks: []*Stack{
				makeStack("alpha", "backend"),
				makeStack("beta", "backend"),
			},
			tag:  "backend",
			want: []string{"alpha", "beta"},
		},
		{
			name: "stack with multiple tags matches on any",
			stacks: []*Stack{
				makeStack("alpha", "backend", "api"),
			},
			tag:  "api",
			want: []string{"alpha"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FilterByTag(tc.stacks, tc.tag)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d stacks, want %d: %v", len(got), len(tc.want), got)
			}
			for i, s := range got {
				if s.Name != tc.want[i] {
					t.Errorf("result[%d].Name = %q, want %q", i, s.Name, tc.want[i])
				}
			}
		})
	}
}

func TestFindByName(t *testing.T) {
	stacks := []*Stack{
		makeStack("alpha"),
		makeStack("beta"),
		makeStack("gamma"),
	}

	tests := []struct {
		name      string
		stackList []*Stack
		query     string
		wantNil   bool
		wantName  string
	}{
		{
			name:     "found returns correct stack",
			stackList: stacks,
			query:    "beta",
			wantName: "beta",
		},
		{
			name:      "not found returns nil",
			stackList: stacks,
			query:     "delta",
			wantNil:   true,
		},
		{
			name:      "empty slice returns nil",
			stackList: nil,
			query:     "alpha",
			wantNil:   true,
		},
		{
			name:     "first element found",
			stackList: stacks,
			query:    "alpha",
			wantName: "alpha",
		},
		{
			name:     "last element found",
			stackList: stacks,
			query:    "gamma",
			wantName: "gamma",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FindByName(tc.stackList, tc.query)
			if tc.wantNil {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("got nil, want stack with name %q", tc.wantName)
			}
			if got.Name != tc.wantName {
				t.Errorf("got name %q, want %q", got.Name, tc.wantName)
			}
		})
	}
}
