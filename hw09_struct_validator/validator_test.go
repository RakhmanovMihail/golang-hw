package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

type ValidatorTestSuite struct {
	suite.Suite
}

func TestValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}

func (s *ValidatorTestSuite) TestValidate() {
	tests := []struct {
		name          string
		input         interface{}
		expectedError ValidationErrors
	}{
		{
			name: "valid user",
			input: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Name:   "Test User",
				Age:    25,
				Email:  "user@example.com",
				Role:   "admin",
				Phones: []string{"79001234567", "79007654321"},
			},
			expectedError: nil,
		},
		{
			name: "invalid user ID length",
			input: User{
				ID:     "123",
				Name:   "Test User",
				Age:    25,
				Email:  "user@example.com",
				Role:   "admin",
				Phones: []string{"79001234567"},
			},
			expectedError: ValidationErrors{
				{Field: "ID", Err: errors.New("field length must be 36")},
			},
		},
		{
			name: "invalid user age (too young)",
			input: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Name:   "Test User",
				Age:    16,
				Email:  "user@example.com",
				Role:   "admin",
				Phones: []string{"79001234567"},
			},
			expectedError: ValidationErrors{
				{Field: "Age", Err: errors.New("value must be at least 18")},
			},
		},
		{
			name: "invalid email format",
			input: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Name:   "Test User",
				Age:    25,
				Email:  "invalid-email",
				Role:   "admin",
				Phones: []string{"79001234567"},
			},
			expectedError: ValidationErrors{
				{Field: "Email", Err: errors.New("value does not match the pattern")},
			},
		},
		{
			name: "invalid phone number length",
			input: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Name:   "Test User",
				Age:    25,
				Email:  "user@example.com",
				Role:   "admin",
				Phones: []string{"12345"},
			},
			expectedError: ValidationErrors{
				{Field: "Phones", Err: errors.New("element 0: field length must be 11")},
			},
		},
		{
			name: "valid app version",
			input: App{
				Version: "1.0.0",
			},
		},
		{
			name: "invalid app version length",
			input: App{
				Version: "1.00.00",
			},
			expectedError: ValidationErrors{
				{Field: "Version", Err: errors.New("field length must be 5")},
			},
		},
		{
			name: "valid response code",
			input: Response{
				Code: 200,
				Body: "OK",
			},
		},
		{
			name: "invalid response code",
			input: Response{
				Code: 400,
				Body: "Bad Request",
			},
			expectedError: ValidationErrors{
				{Field: "Code", Err: errors.New("value must be one of: 200, 404, 500")},
			},
		},
		{
			name:  "non-struct input",
			input: "just a string",
			expectedError: ValidationErrors{
				{Field: "Struct", Err: errors.New("expected a struct")},
			},
		},
		{
			name: "valid token with no validation rules",
			input: Token{
				Header:    []byte("header"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if tt.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, errors.New(tt.expectedError.Error()), err)
			}
		})
	}
}
