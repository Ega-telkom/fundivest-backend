// internal/pkg/validator/validator.go
package validator

import (
    "fmt"
    "strings"
    "github.com/google/uuid"
)

func IsValidUUID(s string) bool {
    _, err := uuid.Parse(s)
    return err == nil
}

func IsValidChapter(chapter int) error {
    if chapter < 1 || chapter > 3 {
        return fmt.Errorf("chapter must be between 1 and 3")
    }
    return nil
}

func IsValidName(name string) error {
    trimmed := strings.TrimSpace(name)
    if len(trimmed) < 2 {
        return fmt.Errorf("name must be at least 2 characters")
    }
    if len(name) > 100 {
        return fmt.Errorf("name too long (max 100 characters)")
    }
    return nil
}