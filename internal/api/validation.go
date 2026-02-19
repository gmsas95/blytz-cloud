package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Validation errors
var (
	ErrInstructionsTooLong   = errors.New("custom_instructions exceeds maximum length of 5000 characters")
	ErrAssistantNameTooLong  = errors.New("assistant_name exceeds maximum length of 50 characters")
	ErrInvalidBotTokenFormat = errors.New("telegram_bot_token format should be: <numbers>:<alphanumeric>")
)

// Validator defines the interface for request validators
type Validator interface {
	Validate(req *CreateCustomerRequest) error
}

// ValidatorFunc is an adapter to allow ordinary functions as validators
type ValidatorFunc func(req *CreateCustomerRequest) error

func (f ValidatorFunc) Validate(req *CreateCustomerRequest) error {
	return f(req)
}

// CompositeValidator runs multiple validators
type CompositeValidator struct {
	validators []Validator
}

// NewCompositeValidator creates a new composite validator
func NewCompositeValidator(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{validators: validators}
}

// Validate runs all validators and returns the first error
func (cv *CompositeValidator) Validate(req *CreateCustomerRequest) error {
	for _, v := range cv.validators {
		if err := v.Validate(req); err != nil {
			return err
		}
	}
	return nil
}

// InstructionsLengthValidator validates custom instructions length
type InstructionsLengthValidator struct {
	MaxLength int
}

// Validate checks if custom instructions exceed maximum length
func (v *InstructionsLengthValidator) Validate(req *CreateCustomerRequest) error {
	if len(req.CustomInstructions) > v.MaxLength {
		return ErrInstructionsTooLong
	}
	return nil
}

// AssistantNameLengthValidator validates assistant name length
type AssistantNameLengthValidator struct {
	MaxLength int
}

// Validate checks if assistant name exceeds maximum length
func (v *AssistantNameLengthValidator) Validate(req *CreateCustomerRequest) error {
	if len(req.AssistantName) > v.MaxLength {
		return ErrAssistantNameTooLong
	}
	return nil
}

// BotTokenFormatValidator validates Telegram bot token format
type BotTokenFormatValidator struct{}

// Validate checks if bot token has the correct format
func (v *BotTokenFormatValidator) Validate(req *CreateCustomerRequest) error {
	if !strings.Contains(req.TelegramBotToken, ":") {
		return ErrInvalidBotTokenFormat
	}
	return nil
}

// DefaultValidators returns the standard set of validators
func DefaultValidators() *CompositeValidator {
	return NewCompositeValidator(
		&InstructionsLengthValidator{MaxLength: 5000},
		&AssistantNameLengthValidator{MaxLength: 50},
		&BotTokenFormatValidator{},
	)
}

// ValidationMiddleware creates a Gin middleware for request validation
func ValidationMiddleware(validator Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateCustomerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "validation_failed",
				Message: "Invalid request body: " + err.Error(),
			})
			c.Abort()
			return
		}

		if err := validator.Validate(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "validation_failed",
				Message: err.Error(),
			})
			c.Abort()
			return
		}

		// Store validated request in context for handler
		c.Set("validated_request", &req)
		c.Next()
	}
}

// GetValidatedRequest retrieves the validated request from context
func GetValidatedRequest(c *gin.Context) *CreateCustomerRequest {
	req, exists := c.Get("validated_request")
	if !exists {
		return nil
	}
	return req.(*CreateCustomerRequest)
}
