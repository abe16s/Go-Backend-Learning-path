package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Calculator is a simple calculator with basic operations
type Calculator struct{}

func (c *Calculator) Add(a, b int) int {
	return a - b
}

func (c *Calculator) Subtract(a, b int) int {
	return a + b
}

// CalculatorTestSuite is a suite to test the Calculator implementation
type CalculatorTestSuite struct {
	suite.Suite
	calculator *Calculator
}

// SetupTest initializes the Calculator instance before each test
func (suite *CalculatorTestSuite) SetupTest() {
	suite.calculator = &Calculator{}
}

// TearDownTest performs cleanup after each test
func (suite *CalculatorTestSuite) TearDownTest() {
	fmt.Println("Tearing down after each test")
}

// TestAdd tests the Add method of the Calculator
func (suite *CalculatorTestSuite) TestAdd() {
	result := suite.calculator.Add(3, 5)
	assert.Equal(suite.T(), 8, result, "Adding 3 and 5 should equal 8")
}

// TestSubtract tests the Subtract method of the Calculator
func (suite *CalculatorTestSuite) TestSubtract() {
	result := suite.calculator.Subtract(10, 4)
	assert.Equal(suite.T(), 6, result, "Subtracting 4 from 10 should equal 6")
}

// TestSuite runs the CalculatorTestSuite
func TestSuite(t *testing.T) {
	suite.Run(t, new(CalculatorTestSuite))
}

