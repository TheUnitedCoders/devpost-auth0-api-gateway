package slice

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertFuncWithSkip(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		slice     []int
		want      []string
		skipElems map[int]struct{}
	}
	tests := []testCase{
		{
			name: "Simple",
			slice: []int{
				1, 2, 3,
			},
			want: []string{
				"1", "3",
			},
			skipElems: map[int]struct{}{
				2: {},
			},
		},
		{
			name:  "Nil",
			slice: nil,
			want:  nil,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				tt.want,
				ConvertFuncWithSkip(tt.slice,
					func(elem int) (string, bool) {
						_, skip := tt.skipElems[elem]
						return strconv.Itoa(elem), skip
					},
				),
			)
		})
	}
}

func TestConvertFunc(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name  string
		slice []int
		want  []string
	}
	tests := []testCase{
		{
			name: "Simple",
			slice: []int{
				1, 2, 3,
			},
			want: []string{
				"1", "2", "3",
			},
		},
		{
			name:  "Nil",
			slice: nil,
			want:  nil,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				tt.want,
				ConvertFunc(tt.slice, strconv.Itoa),
			)
		})
	}
}

func TestMerge(t *testing.T) {
	t.Parallel()

	type args struct {
		s1 []int
		s2 []int
	}
	type testCase struct {
		name string
		args args
		want []int
	}
	tests := []testCase{
		{
			name: "With deduplication",
			args: args{
				s1: []int{1, 2, 3},
				s2: []int{3, 2, 1},
			},
			want: []int{1, 2, 3},
		},
		{
			name: "Without deduplication",
			args: args{
				s1: []int{1, 2, 3},
				s2: []int{4, 5, 6},
			},
			want: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name: "First slice is nil",
			args: args{
				s1: nil,
				s2: []int{4, 5, 6},
			},
			want: []int{4, 5, 6},
		},
		{
			name: "Second slice is nil",
			args: args{
				s1: []int{1, 2, 3},
				s2: nil,
			},
			want: []int{1, 2, 3},
		},
		{
			name: "Both slices is nil",
			args: args{
				s1: nil,
				s2: nil,
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.ElementsMatch(t, tt.want, Merge(tt.args.s1, tt.args.s2))
		})
	}
}
