package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInstructionsLengthValidator(t *testing.T) {
	validator := &InstructionsLengthValidator{MaxLength: 10}

	tests := []struct {
		name    string
		req     CreateCustomerRequest
		wantErr bool
	}{
		{
			name:    "valid length",
			req:     CreateCustomerRequest{CustomInstructions: "short"},
			wantErr: false,
		},
		{
			name:    "exactly at limit",
			req:     CreateCustomerRequest{CustomInstructions: "0123456789"},
			wantErr: false,
		},
		{
			name:    "over limit",
			req:     CreateCustomerRequest{CustomInstructions: "01234567890"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.req)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInstructionsTooLong)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAssistantNameLengthValidator(t *testing.T) {
	validator := &AssistantNameLengthValidator{MaxLength: 10}

	tests := []struct {
		name    string
		req     CreateCustomerRequest
		wantErr bool
	}{
		{
			name:    "valid length",
			req:     CreateCustomerRequest{AssistantName: "short"},
			wantErr: false,
		},
		{
			name:    "over limit",
			req:     CreateCustomerRequest{AssistantName: "this is way too long"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.req)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrAssistantNameTooLong)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBotTokenFormatValidator(t *testing.T) {
	validator := &BotTokenFormatValidator{}

	tests := []struct {
		name    string
		req     CreateCustomerRequest
		wantErr bool
	}{
		{
			name:    "valid format",
			req:     CreateCustomerRequest{TelegramBotToken: "123456:ABC-DEF"},
			wantErr: false,
		},
		{
			name:    "missing colon",
			req:     CreateCustomerRequest{TelegramBotToken: "123456ABC"},
			wantErr: true,
		},
		{
			name:    "empty string",
			req:     CreateCustomerRequest{TelegramBotToken: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.req)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInvalidBotTokenFormat)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatorFunc(t *testing.T) {
	called := false
	fn := ValidatorFunc(func(req *CreateCustomerRequest) error {
		called = true
		return nil
	})

	err := fn.Validate(&CreateCustomerRequest{Email: "test@test.com"})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestNewCompositeValidator(t *testing.T) {
	v1 := &InstructionsLengthValidator{MaxLength: 100}
	v2 := &AssistantNameLengthValidator{MaxLength: 50}

	cv := NewCompositeValidator(v1, v2)
	assert.NotNil(t, cv)
	assert.Len(t, cv.validators, 2)

	// Test with empty validators
	emptyCV := NewCompositeValidator()
	assert.NotNil(t, emptyCV)
	assert.Len(t, emptyCV.validators, 0)

	// Empty validator should always pass
	err := emptyCV.Validate(&CreateCustomerRequest{})
	assert.NoError(t, err)
}

func TestCompositeValidator(t *testing.T) {
	validator := NewCompositeValidator(
		&InstructionsLengthValidator{MaxLength: 5000},
		&AssistantNameLengthValidator{MaxLength: 50},
		&BotTokenFormatValidator{},
	)

	tests := []struct {
		name    string
		req     CreateCustomerRequest
		wantErr bool
	}{
		{
			name: "all valid",
			req: CreateCustomerRequest{
				Email:              "test@example.com",
				AssistantName:      "Test Assistant",
				CustomInstructions: "Help me",
				TelegramBotToken:   "123456:ABC-DEF",
			},
			wantErr: false,
		},
		{
			name: "instructions too long",
			req: CreateCustomerRequest{
				CustomInstructions: string(make([]byte, 5001)),
			},
			wantErr: true,
		},
		{
			name: "assistant name too long",
			req: CreateCustomerRequest{
				AssistantName: string(make([]byte, 51)),
			},
			wantErr: true,
		},
		{
			name: "invalid bot token",
			req: CreateCustomerRequest{
				TelegramBotToken: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(&tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name: "valid request",
			body: CreateCustomerRequest{
				Email:              "test@example.com",
				AssistantName:      "Test",
				CustomInstructions: "Help",
				TelegramBotToken:   "123:ABC",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "validation fails",
			body: CreateCustomerRequest{
				TelegramBotToken: "invalid",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/test", ValidationMiddleware(DefaultValidators()), func(c *gin.Context) {
				req := GetValidatedRequest(c)
				if req != nil {
					c.JSON(http.StatusOK, req)
				} else {
					c.Status(http.StatusInternalServerError)
				}
			})

			var body []byte
			var err error
			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.body)
				assert.NoError(t, err)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetValidatedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/test", ValidationMiddleware(DefaultValidators()), func(c *gin.Context) {
		req := GetValidatedRequest(c)
		if req != nil {
			c.JSON(http.StatusOK, gin.H{"email": req.Email})
		} else {
			c.Status(http.StatusInternalServerError)
		}
	})

	body, _ := json.Marshal(CreateCustomerRequest{
		Email:              "test@example.com",
		AssistantName:      "Test",
		CustomInstructions: "Help",
		TelegramBotToken:   "123:ABC",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetValidatedRequestNotInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		req := GetValidatedRequest(c)
		if req == nil {
			c.JSON(http.StatusOK, gin.H{"result": "nil"})
		} else {
			c.JSON(http.StatusOK, gin.H{"result": "not nil"})
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "nil", response["result"])
}
