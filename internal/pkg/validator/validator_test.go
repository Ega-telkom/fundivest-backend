// internal/pkg/validator/validator_test.go
package validator_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "fundivest/internal/pkg/validator"
)

func TestIsValidUUID(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  bool
    }{
        {"valid uuid", "550e8400-e29b-41d4-a716-446655440000", true},
        {"invalid uuid", "not-a-uuid", false},
        {"empty string", "", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.want, validator.IsValidUUID(tt.input))
        })
    }
}

func TestIsValidName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid name", "John Doe", false},
        {"too short", "A", true},
        {"empty", "", true},
        {"too long", string(make([]byte, 101)), true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.IsValidName(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}