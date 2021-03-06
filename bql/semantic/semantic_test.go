// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package semantic

import (
	"reflect"
	"testing"

	"github.com/google/badwolf/triple"
	"github.com/google/badwolf/triple/literal"
	"github.com/google/badwolf/triple/node"
	"github.com/google/badwolf/triple/predicate"
)

func TestStatementType(t *testing.T) {
	st := &Statement{}
	st.BindType(Query)
	if got, want := st.Type(), Query; got != want {
		t.Errorf("semantic.NewStatement returned wrong statement type; got %s, want %s", got, want)
	}
}

func TestStatementAddGraph(t *testing.T) {
	st := &Statement{}
	st.BindType(Query)
	st.AddGraph("?foo")
	if got, want := st.GraphNames(), []string{"?foo"}; !reflect.DeepEqual(got, want) {
		t.Errorf("semantic.AddGraph returned the wrong graphs available; got %v, want %v", got, want)
	}
}

func TestStatementAddData(t *testing.T) {
	tr, err := triple.Parse(`/_<foo> "foo"@[] /_<bar>`, literal.DefaultBuilder())
	if err != nil {
		t.Fatalf("triple.Parse failed to parse valid triple with error %v", err)
	}
	st := &Statement{}
	st.BindType(Query)
	st.AddData(tr)
	if got, want := st.Data(), []*triple.Triple{tr}; !reflect.DeepEqual(got, want) {
		t.Errorf("semantic.AddData returned the wrong data available; got %v, want %v", got, want)
	}
}

func TestGraphClauseSpecificity(t *testing.T) {
	table := []struct {
		gc   *GraphClause
		want int
	}{
		{&GraphClause{}, 0},
		{&GraphClause{S: &node.Node{}}, 1},
		{&GraphClause{S: &node.Node{}, P: &predicate.Predicate{}}, 2},
		{&GraphClause{S: &node.Node{}, P: &predicate.Predicate{}, O: &triple.Object{}}, 3},
	}
	for _, entry := range table {
		if got, want := entry.gc.Specificity(), entry.want; got != want {
			t.Errorf("semantic.GraphClause.Specificity failed to return the proper value for %v; got %d, want %d", entry.gc, got, want)
		}
	}
}

func TestGraphClauseManipulation(t *testing.T) {
	st := &Statement{}
	if st.WorkingClause() != nil {
		t.Fatalf("semantic.GraphClause.WorkingClause should not return a working clause without initilization in %v", st)
	}
	st.ResetWorkingGraphClause()
	if st.WorkingClause() == nil {
		t.Fatalf("semantic.GraphClause.WorkingClause should return a working clause after initilization in %v", st)
	}
	st.AddWorkingGraphClause()
	if got, want := len(st.GraphPatternClauses()), 0; got != want {
		t.Fatalf("semantic.GraphClause.Clauses return wrong number of clauses in %v; got %d, want %d", st, got, want)
	}
}

func TestBindingListing(t *testing.T) {
	stm := Statement{}
	stm.ResetWorkingGraphClause()
	for i := 0; i < 10; i++ {
		wcls := stm.WorkingClause()
		v := string(i)
		cls := &GraphClause{
			SBinding:         "?" + v,
			SAlias:           "?" + v,
			STypeAlias:       "?" + v,
			SIDAlias:         "?" + v,
			PAlias:           "?" + v,
			PID:              "?" + v,
			PAnchorBinding:   "?" + v,
			PBinding:         "?" + v,
			PLowerBoundAlias: "?" + v,
			PUpperBoundAlias: "?" + v,
			PIDAlias:         "?" + v,
			PAnchorAlias:     "?" + v,
			OBinding:         "?" + v,
			OID:              "?" + v,
			OAlias:           "?" + v,
			OTypeAlias:       "?" + v,
			OIDAlias:         "?" + v,
			OAnchorAlias:     "?" + v,
			OAnchorBinding:   "?" + v,
			OLowerBoundAlias: "?" + v,
			OUpperBoundAlias: "?" + v,
		}
		*wcls = *cls
		stm.AddWorkingGraphClause()
	}
	bds := stm.BindingsMap()
	if len(bds) != 10 {
		t.Errorf("Statement.Bindings failed to reteurn 10 bindings, instead returned %v", bds)
	}
	for b, cnt := range bds {
		if cnt != 19 {
			t.Errorf("Statement.Bindings failed to update binding %q to 20, got %d instead", b, cnt)
		}
	}
}

func TestIsEmptyClause(t *testing.T) {
	testTable := []struct {
		in  *GraphClause
		out bool
	}{
		{
			in:  &GraphClause{},
			out: true,
		},
		{
			in:  &GraphClause{SBinding: "?foo"},
			out: false,
		},
	}
	for _, entry := range testTable {
		if got, want := entry.in.IsEmpty(), entry.out; got != want {
			t.Errorf("IsEmpty for %v returned %v, but should have returned %v", entry.in, got, want)
		}
	}

}

func TestSortedGraphPatternClauses(t *testing.T) {
	s := &Statement{
		pattern: []*GraphClause{
			{},
			{S: &node.Node{}},
			{S: &node.Node{}, P: &predicate.Predicate{}},
			{S: &node.Node{}, P: &predicate.Predicate{}, O: &triple.Object{}},
		},
	}
	spc := 3
	for _, cls := range s.SortedGraphPatternClauses() {
		if want, got := spc, cls.Specificity(); got != want {
			t.Errorf("statement.SortedGraphPatternClauses failed to sort properly; got specificity %d, want specificity %d", got, want)
		}
		spc--
	}
}

func TestProjectionIsEmpty(t *testing.T) {
	s := &Statement{}
	s.ResetProjection()
	if !s.WorkingProjection().IsEmpty() {
		t.Errorf("s.WorkingProjections().IsEmpty() should be empty after reset, instead got %s", s.WorkingProjection())
	}
	if len(s.Projections()) != 0 {
		t.Errorf("s.Projections should be empty, instead got %s", s.Projections())
	}
	s.AddWorkingProjection()
	if len(s.Projections()) != 0 {
		t.Errorf("s.Projections should be empty after adding an empty projection, instead got %s", s.Projections())
	}
	p := s.WorkingProjection()
	p.Binding = "?foo"
	s.AddWorkingProjection()
	if len(s.Projections()) != 1 {
		t.Errorf("s.Projections should constina one projection, instead got %s", s.Projections())
	}
}

func TestConstructClauseManipulation(t *testing.T) {
	st := &Statement{}
	if st.WorkingConstructClause() != nil {
		t.Fatalf("semantic.ConstructClause.WorkingConstructClause should not return a working construct clause without initialization in %v", st)
	}
	st.ResetWorkingConstructClause()
	if st.WorkingConstructClause() == nil {
		t.Fatalf("semantic.ConstructClause.WorkingConstructClause should return a working construct clause after initialization in %v", st)
	}
	st.AddWorkingConstructClause()
	if got, want := len(st.ConstructClauses()), 0; got != want {
		t.Fatalf("semantic.ConstructClause.ConstructClauses returns wrong number of clauses in %v; got %d, want %d", st, got, want)
	}
}

func TestReificationClauseManipulation(t *testing.T) {
	st := &Statement{}
	st.ResetWorkingConstructClause()
	wcc := st.WorkingConstructClause()
	if wcc.WorkingReificationClause() != nil {
		t.Fatalf("semantic.ConstructClause.WorkingReificationClause should not return a working reification clause without initialization in %v", st)
	}
	wcc.ResetWorkingReificationClause()
	if wcc.WorkingReificationClause() == nil {
		t.Fatalf("semantic.ConstructClause.WorkingReificationClause should return a working reification clause after initialization in %v", st)
	}
	wcc.AddWorkingReificationClause()
	if got, want := len(wcc.ReificationClauses()), 0; got != want {
		t.Fatalf("semantic.ConstructClause.WorkingReificationClauses returns wrong number of clauses in %v; got %d, want %d", st, got, want)
	}
}

func TestInputOutputBindings(t *testing.T) {
	s := &Statement{
		projection: []*Projection{
			{
				Binding: "?foo",
				Alias:   "?foo_alias",
			},
			{
				Binding: "?bar",
			},
		},
		constructClauses: []*ConstructClause{
			{
				SBinding: "?foo1",
				PBinding: "?foo2",
				OBinding: "?foo3",
			},
			{
				SBinding: "?foo4",
				PBinding: "?foo5",
				OBinding: "?foo6",
				reificationClauses: []*ReificationClause{
					{
						PBinding: "?foo7",
						OBinding: "?foo8",
					},
					{
						PBinding: "?foo9",
						OBinding: "?foo10",
					},

				},
			},
			{
				PAnchorBinding: "?foo11",
				OAnchorBinding: "?foo12",
				reificationClauses: []*ReificationClause{
					{
						PAnchorBinding: "?foo13",
						OAnchorBinding: "?foo14",
					},

				},
			},
		},
	}
	want := []string{"?foo", "?bar", "?foo1", "?foo2", "?foo3", "?foo4", "?foo5", "?foo6",
	        "?foo7","?foo8", "?foo9", "?foo10", "?foo11", "?foo12", "?foo13", "?foo14"}
	if got := s.InputBindings(); !reflect.DeepEqual(got, want) {
		t.Errorf("s.InputBindings return the wrong input binding; got %v, want %v", got, want)
	}
	if got, want := s.OutputBindings(), []string{"?foo_alias", "?bar"}; !reflect.DeepEqual(got, want) {
		t.Errorf("s.OutputBindings return the wrong input binding; got %v, want %v", got, want)
	}
}
