package main

import (
	"fmt"
	"github.com/infisical/gocasl"
	"log"
)

// 1. Define Subject
type Product struct {
	ID    string
	Price float64
	Tags  []string
}

func (Product) SubjectType() string {
	return "Product"
}

func (p Product) GetField(field string) any {
	switch field {
	case "ID":
		return p.ID
	case "Price":
		return p.Price
	case "Tags":
		return p.Tags
	default:
		return nil
	}
}

// 2. Define Custom Operator Function
// This function checks if the fieldValue (number) is between [min, max] inclusive.
func opBetween(fieldValue, operand any) bool {
	// 1. Cast field value to float64 (assuming standard number type for this example)
	val, ok := toFloat(fieldValue)
	if !ok {
		return false
	}

	// 2. Cast operand to array/slice
	bounds, ok := operand.([]any)
	if !ok || len(bounds) != 2 {
		return false
	}

	min, minOk := toFloat(bounds[0])
	max, maxOk := toFloat(bounds[1])

	if !minOk || !maxOk {
		return false
	}

	return val >= min && val <= max
}

// Helper to convert any number to float64
func toFloat(v any) (float64, bool) {
	switch i := v.(type) {
	case float64:
		return i, true
	case float32:
		return float64(i), true
	case int:
		return float64(i), true
	case int64:
		return float64(i), true
	default:
		return 0, false
	}
}

// 3. Define DSL Helper (Optional, but nice for type safety in Go code)
func Between(min, max float64) gocasl.Op {
	return gocasl.Op{"$between": []any{min, max}}
}

func main() {
	// Define Actions
	read := gocasl.DefineAction[Product]("read")
	buy := gocasl.DefineAction[Product]("buy")

	// 4. Register Operator in AbilityBuilder
	// Start with DefaultOperators and add ours
	myOperators := gocasl.DefaultOperators().With("$between", opBetween)

	builder := gocasl.NewAbility().WithOperators(myOperators)

	// Add Rules using the custom operator
	// Allow reading any product
	gocasl.AddRule(builder, gocasl.Allow(read).Build())

	// Allow buying only products with price between 10 and 100
	gocasl.AddRule(builder, gocasl.Allow(buy).
		Where(gocasl.Cond{
			"Price": Between(10, 100),
		}).
		Build())

	ability, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}

	// 5. Test It
	cheap := Product{ID: "p1", Price: 5.0}
	affordable := Product{ID: "p2", Price: 50.0}
	expensive := Product{ID: "p3", Price: 200.0}

	fmt.Println("Checking Permissions (Price Policy: $10 - $100)")

	fmt.Printf("Can buy Cheap ($5.0)? %v\n", gocasl.Can(ability, buy, cheap))            // false
	fmt.Printf("Can buy Affordable ($50.0)? %v\n", gocasl.Can(ability, buy, affordable)) // true
	fmt.Printf("Can buy Expensive ($200.0)? %v\n", gocasl.Can(ability, buy, expensive))  // false

	// Check read (always true)
	fmt.Printf("Can read Expensive? %v\n", gocasl.Can(ability, read, expensive)) // true
}
