# Domain Layer

This package contains the core business domain models and interfaces.

## Purpose

The domain layer represents the heart of the business logic, containing:
- Core entities and value objects
- Business rules and invariants
- Domain services interfaces
- Domain events

## Guidelines

1. **Keep it Pure**: Domain models should have no dependencies on external packages
2. **Rich Models**: Encapsulate business logic within domain models
3. **Immutability**: Prefer immutable value objects where appropriate
4. **Validation**: Domain models should validate their own state

## Example Structure

```go
// user.go
package domain

import (
    "errors"
    "time"
)

// User represents a user entity
type User struct {
    ID        string
    Email     Email    // Value object
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// NewUser creates a new user with validation
func NewUser(email string, name string) (*User, error) {
    emailVO, err := NewEmail(email)
    if err != nil {
        return nil, err
    }
    
    if name == "" {
        return nil, errors.New("name cannot be empty")
    }
    
    return &User{
        ID:        generateID(),
        Email:     emailVO,
        Name:      name,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}

// Email is a value object
type Email struct {
    value string
}

// NewEmail creates a validated email
func NewEmail(email string) (Email, error) {
    if !isValidEmail(email) {
        return Email{}, errors.New("invalid email format")
    }
    return Email{value: email}, nil
}

// String returns the email as string
func (e Email) String() string {
    return e.value
}
```

## Repository Interfaces

Define repository interfaces in the domain layer:

```go
// user_repository.go
package domain

import "context"

// UserRepository defines the interface for user persistence
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*User, error)
    GetByEmail(ctx context.Context, email Email) (*User, error)
    Save(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}
```

## Domain Services

For complex business logic that doesn't fit in a single entity:

```go
// user_service.go
package domain

// UserService handles complex user operations
type UserService interface {
    Register(ctx context.Context, email string, name string) (*User, error)
    Authenticate(ctx context.Context, email string, password string) (*User, error)
    ChangeEmail(ctx context.Context, userID string, newEmail string) error
}
```

## Best Practices

1. **Use Value Objects**: For concepts with no identity (Email, Money, Address)
2. **Aggregate Roots**: Define clear aggregate boundaries
3. **Domain Events**: Emit events for significant business occurrences
4. **Specifications**: Use specification pattern for complex business rules
5. **Factory Methods**: Use factory methods for complex object creation
