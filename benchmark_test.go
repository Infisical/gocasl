package gocasl

import (
	"testing"
)

func BenchmarkCan_Simple(b *testing.B) {
	read := DefineAction[mockSubject]("read")
	sub := mockSubject{ID: 1}
	
	builder := NewAbility()
	AddRule(builder, Allow(read).Where(Cond{"ID": 1}).Build())
	a := builder.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Can(a, read, sub)
	}
}

func BenchmarkCan_NoRules(b *testing.B) {
	read := DefineAction[mockSubject]("read")
	sub := mockSubject{ID: 1}
	
	builder := NewAbility()
	a := builder.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Can(a, read, sub)
	}
}

func BenchmarkCan_Complex(b *testing.B) {
	read := DefineAction[mockSubject]("read")
	sub := mockSubject{ID: 50, Title: "Match", Tags: []string{"go"}}
	
	builder := NewAbility()
	
	// Add 100 rules that don't match
	for i := 0; i < 100; i++ {
		AddRule(builder, Allow(read).Where(Cond{"ID": 1000 + i}).Build())
	}
	
	// Add matching rule with complex condition
	AddRule(builder, Allow(read).Where(And(
		Cond{"ID": Op{"$gt": 10}},
		Cond{"Title": "Match"},
		Cond{"Tags": Op{"$contains": "go"}},
	)).Build())
	
	a := builder.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Can(a, read, sub)
	}
}

func BenchmarkCan_LargeRuleSet(b *testing.B) {
	read := DefineAction[mockSubject]("read")
	sub := mockSubject{ID: 5000}
	
	builder := NewAbility()
	
	// Add 10,000 rules
	for i := 0; i < 10000; i++ {
		AddRule(builder, Allow(read).Where(Cond{"ID": i}).Build())
	}
	
	a := builder.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Can(a, read, sub)
	}
}

func BenchmarkRuleIndexing(b *testing.B) {
	read := DefineAction[mockSubject]("read")
	
	rules := make([]Rule[mockSubject], 1000)
	for i := 0; i < 1000; i++ {
		rules[i] = Allow(read).Where(Cond{"ID": i}).Build()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewAbility()
		AddRules(builder, rules...)
		builder.Build()
	}
}
