package handler_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/handler"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/handler/handlerfakes"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/model"
	"github.com/georgisomnoev/feature-flag-api/internal/validator"
	"github.com/labstack/echo/v4"
)

var (
	ErrInvalidOrExpiredToken = errors.New("invalid or expired token")
	ErrInternalError         = errors.New("internal error")
)

var _ = Describe("Handler", func() {
	var (
		e           *echo.Echo
		recorder    *httptest.ResponseRecorder
		authStore   *handlerfakes.FakeAuthStore
		jwtHelper   *handlerfakes.FakeJWTHelper
		svc         *handlerfakes.FakeService
		flagHandler *handler.Handler
		request     *http.Request

		validUserID = "c9c15117-ca25-49c6-b857-3eb640a61234"
	)

	BeforeEach(func() {
		e = echo.New()
		e.Validator = validator.GetValidator()
		recorder = httptest.NewRecorder()
		authStore = &handlerfakes.FakeAuthStore{}
		jwtHelper = &handlerfakes.FakeJWTHelper{}
		svc = &handlerfakes.FakeService{}
		flagHandler = handler.NewHandler(svc, authStore, jwtHelper)
		flagHandler.RegisterHandlers(e)
	})

	JustBeforeEach(func() {
		request = httptest.NewRequest(http.MethodGet, "/flags", nil)
		request.Header.Set("Content-Type", "application/json")
	})

	When("the authorization header is missing", func() {
		It("returns unauthorized error", func() {
			e.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(ContainSubstring("missing token"))
		})
	})

	When("the token is invalid or malformed", func() {
		JustBeforeEach(func() {
			request.Header.Set("Authorization", "InvalidToken")
		})

		BeforeEach(func() {
			jwtHelper.ValidateTokenReturns(nil, ErrInvalidOrExpiredToken)
		})

		It("returns unauthorized error", func() {
			e.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			Expect(recorder.Body.String()).To(ContainSubstring(ErrInvalidOrExpiredToken.Error()))
		})
	})

	When("a token is provided ", func() {
		JustBeforeEach(func() {
			request.Header.Set("Authorization", "Bearer validToken")
		})

		Context("and contains an invalid or missing user ID", func() {
			BeforeEach(func() {
				claims := jwt.MapClaims{"sub": ""}
				jwtHelper.ValidateTokenReturns(claims, nil)
			})

			It("returns unauthorized error", func() {
				e.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				Expect(recorder.Body.String()).To(ContainSubstring("invalid user ID in token"))
			})
		})

		Context("and contains valid claims", func() {
			BeforeEach(func() {
				claims := jwt.MapClaims{"sub": validUserID}
				jwtHelper.ValidateTokenReturns(claims, nil)
			})

			Context("but the user does not exist", func() {
				BeforeEach(func() {
					authStore.UserExistsReturns(false, nil)
				})

				It("returns unauthorized error", func() {
					e.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
					Expect(recorder.Body.String()).To(ContainSubstring("user not found"))
				})
			})

			Context("the user exists but the scope is not set", func() {
				BeforeEach(func() {
					authStore.UserExistsReturns(true, nil)
				})

				It("returns internal server error", func() {
					e.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusForbidden))
					Expect(recorder.Body.String()).To(ContainSubstring("no scopes found in token"))
				})

				Context("the scope is set but does not have the required scopes", func() {
					BeforeEach(func() {
						claims := jwt.MapClaims{"sub": validUserID, "scopes": []string{"other:string"}}
						jwtHelper.ValidateTokenReturns(claims, nil)
					})

					It("returns forbidden error", func() {
						e.ServeHTTP(recorder, request)
						Expect(recorder.Code).To(Equal(http.StatusForbidden))
						Expect(recorder.Body.String()).To(ContainSubstring("insufficient permissions"))
					})
				})
			})
		})
	})

	Describe("GET /flags", func() {
		BeforeEach(func() {
			claims := jwt.MapClaims{"sub": validUserID, "scopes": []string{"read:flags"}}
			jwtHelper.ValidateTokenReturns(claims, nil)
			authStore.UserExistsReturns(true, nil)
		})

		JustBeforeEach(func() {
			request.Header.Set("Authorization", "Bearer validToken")
		})

		Context("when the request is successful", func() {
			var featureFlag model.FeatureFlag
			BeforeEach(func() {
				featureFlag = model.FeatureFlag{ID: uuid.New(), Key: "flag1", Description: "desc1", Enabled: true}
				svc.ListFlagsReturns([]model.FeatureFlag{featureFlag}, nil)
			})

			It("returns the list of feature flags", func() {
				e.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(ContainSubstring(featureFlag.Key))
			})
		})

		Context("when the service returns an error", func() {
			BeforeEach(func() {
				svc.ListFlagsReturns(nil, ErrInternalError)
			})

			It("returns an internal server error", func() {
				e.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(ContainSubstring(ErrInternalError.Error()))
			})
		})
	})

	Describe("GET /flags/:id", func() {
		var (
			flagIDStr string
			flagID    uuid.UUID
		)

		BeforeEach(func() {
			claims := jwt.MapClaims{"sub": validUserID, "scopes": []string{"read:flags"}}
			jwtHelper.ValidateTokenReturns(claims, nil)
			authStore.UserExistsReturns(true, nil)
			flagIDStr = "123e4567-e89b-12d3-a456-426655440000"
			flagID = uuid.MustParse(flagIDStr)
		})

		JustBeforeEach(func() {
			request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/flags/%s", flagIDStr), nil)
			request.Header.Set("Authorization", "Bearer validToken")
		})

		Context("when the request is successful", func() {
			BeforeEach(func() {
				svc.GetFlagByIDReturns(model.FeatureFlag{ID: flagID, Key: "flag1", Description: "desc1", Enabled: true}, nil)
			})

			It("returns the feature flag", func() {
				e.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(ContainSubstring("flag1"))
			})
		})

		Context("when the feature flag is not found", func() {
			BeforeEach(func() {
				svc.GetFlagByIDReturns(model.FeatureFlag{}, model.ErrNotFound)
			})

			It("returns a not found error", func() {
				e.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(ContainSubstring("feature flag not found"))
			})
		})

		Context("when there is an invalid flag ID", func() {
			BeforeEach(func() {
				flagIDStr = "invalid-id"
			})

			It("returns a bad request error", func() {
				e.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(ContainSubstring("invalid flag ID"))
			})
		})
	})

	Describe("POST /flags", func() {
		var payload string
		BeforeEach(func() {
			claims := jwt.MapClaims{"sub": validUserID, "scopes": []string{"write:flags"}}
			jwtHelper.ValidateTokenReturns(claims, nil)
			authStore.UserExistsReturns(true, nil)

			payload = `{"key":"new-flag", "description":"new flag description", "enabled":true}`
		})

		JustBeforeEach(func() {
			request = httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(payload))
			request.Header.Set("Authorization", "Bearer validToken")
			request.Header.Set("Content-Type", "application/json")
		})

		It("succeeds", func() {
			e.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusCreated))
		})

		Context("when the payload is invalid", func() {
			BeforeEach(func() {
				payload = `{"description":"missing key", "enabled":true}`
			})

			It("returns a bad request error", func() {
				e.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("PUT /flags/:id", func() {
		var payload string
		BeforeEach(func() {
			claims := jwt.MapClaims{"sub": validUserID, "scopes": []string{"write:flags"}}
			jwtHelper.ValidateTokenReturns(claims, nil)
			authStore.UserExistsReturns(true, nil)

			payload = `{"key":"updated-flag", "description":"updated desc", "enabled":false}`
		})

		JustBeforeEach(func() {
			request = httptest.NewRequest(http.MethodPut, "/flags/123e4567-e89b-12d3-a456-426655440000", strings.NewReader(payload))
			request.Header.Set("Authorization", "Bearer validToken")
			request.Header.Set("Content-Type", "application/json")
		})

		It("succeeds", func() {
			e.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		Context("when the flag is not found", func() {
			BeforeEach(func() {
				svc.UpdateFlagReturns(model.ErrNotFound)
			})

			It("returns not found error", func() {
				e.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(ContainSubstring(model.ErrNotFound.Error()))
			})
		})

		Context("when the payload is invalid", func() {
			BeforeEach(func() {
				payload = `{"invalid_field":"value"}`
			})

			It("returns bad request error", func() {
				e.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("DELETE /flags/:id", func() {
		var flagIDStr string
		BeforeEach(func() {
			flagIDStr = "123e4567-e89b-12d3-a456-426655440000"
			claims := jwt.MapClaims{"sub": validUserID, "scopes": []string{"write:flags"}}
			jwtHelper.ValidateTokenReturns(claims, nil)
			authStore.UserExistsReturns(true, nil)
		})

		JustBeforeEach(func() {
			request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/flags/%s", flagIDStr), nil)
			request.Header.Set("Authorization", "Bearer validToken")
		})

		It("succeeds", func() {
			e.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusNoContent))
		})

		Context("when the flag is not found", func() {
			BeforeEach(func() {
				svc.DeleteFlagReturns(model.ErrNotFound)
			})

			It("returns not found error", func() {
				e.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(ContainSubstring("feature flag not found"))
			})
		})

		Context("when the ID is invalid", func() {
			BeforeEach(func() {
				flagIDStr = "invalid-id"
			})

			It("returns bad request error", func() {
				e.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})
})
