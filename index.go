package gocasl

import (
	"sync"
)

type ruleIndex struct {
	// Map SubjectType -> Action -> Rules
	index map[string]map[string][]*compiledRule
}

func newRuleIndex(rules []rawRule, compiler *condCompiler) *ruleIndex {
	idx := &ruleIndex{
		index: make(map[string]map[string][]*compiledRule),
	}

	for _, r := range rules {
		idx.add(r, compiler)
	}

	return idx
}

func (i *ruleIndex) add(r rawRule, compiler *condCompiler) {
	if i.index[r.SubjectType] == nil {
		i.index[r.SubjectType] = make(map[string][]*compiledRule)
	}
	
	cr := &compiledRule{
		rule:     r,
		compiler: compiler,
	}
	
	i.index[r.SubjectType][r.Action] = append(i.index[r.SubjectType][r.Action], cr)
}

// get returns rules for a specific subject type and action.
// It returns a slice of compiled rules.
func (i *ruleIndex) get(subjectType string, action string) []*compiledRule {
	if subjectRules, ok := i.index[subjectType]; ok {
		return subjectRules[action]
	}
	return nil
}

type compiledRule struct {
	rule      rawRule
	condition Condition
	compiler  *condCompiler
	once      sync.Once
}

// match evaluates the rule against the subject.
// It lazily compiles the condition on first use.
func (c *compiledRule) match(subject Subject) bool {
	c.once.Do(func() {
		c.condition = c.compiler.compile(c.rule.Conditions)
	})
	return c.condition(subject)
}
