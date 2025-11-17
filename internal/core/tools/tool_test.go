package tools

import (
	"regexp"
	"testing"
)

func TestApplyPathTransformations(t *testing.T) {

	type Test struct {
		Name string
		Pt   []PathTransformation
		In   string
		Want string
	}

	tests := []Test{
		{
			Name: "first match",
			Pt: []PathTransformation{{
				Pattern:     *regexp.MustCompile("^input"),
				Replacement: "replacement",
			}, {
				Pattern:     *regexp.MustCompile("^output"),
				Replacement: "bad_replacement",
			}},
			In:   "input/asd",
			Want: "replacement/asd",
		},
		{
			Name: "second match",
			Pt: []PathTransformation{{
				Pattern:     *regexp.MustCompile("^input"),
				Replacement: "bad_replacement",
			}, {
				Pattern:     *regexp.MustCompile("^output"),
				Replacement: "replacement",
			}},
			In:   "output/asd",
			Want: "replacement/asd",
		},
		{
			Name: "no match",
			Pt: []PathTransformation{{
				Pattern:     *regexp.MustCompile("^input"),
				Replacement: "bad_replacement",
			}, {
				Pattern:     *regexp.MustCompile("^output"),
				Replacement: "bad_replacement",
			}},
			In:   "other/asd",
			Want: "other/asd",
		},
		{
			Name: "only one match",
			Pt: []PathTransformation{{
				Pattern:     *regexp.MustCompile("^input"),
				Replacement: "output",
			}, {
				Pattern:     *regexp.MustCompile("^output"),
				Replacement: "bad_replacement",
			}},
			In:   "input/asd",
			Want: "output/asd",
		},
	}

	for _, tt := range tests {

		t.Run(tt.Name, func(t *testing.T) {
			tool := ToolInstance{
				PathTransformations: tt.Pt,
			}

			got := tool.ApplyPathTransformations(tt.In)

			if tt.Want != got {
				t.Errorf("got %#v, want %#v", got, tt.Want)
			}
		})

	}

}

func TestApply(t *testing.T) {

	type Test struct {
		Name  string
		Pt    PathTransformation
		In    string
		Want1 string
		Want2 bool
	}

	tests := []Test{
		{
			Name: "match",
			Pt: PathTransformation{
				Pattern:     *regexp.MustCompile("^input"),
				Replacement: "replacement",
			},
			In:    "input/asd",
			Want1: "replacement/asd",
			Want2: true,
		},
		{
			Name: "no match",
			Pt: PathTransformation{
				Pattern:     *regexp.MustCompile("^input"),
				Replacement: "replacement",
			},
			In:    "output/asd",
			Want1: "output/asd",
			Want2: false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.Name, func(t *testing.T) {

			got1, got2 := tt.Pt.Apply(tt.In)

			if tt.Want1 != got1 {
				t.Errorf("got %#v, want %#v", got1, tt.Want1)
			} else if tt.Want2 != got2 {
				t.Errorf("got %#v, want %#v", got2, tt.Want2)
			}
		})

	}

}
