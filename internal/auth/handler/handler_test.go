package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/georgisomnoev/feature-flag-api/internal/auth/handler"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/handler/handlerfakes"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/model"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/service"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var ErrUnexpectedError = errors.New("unexpected error")

var _ = Describe("Authentication Handler", func() {
	var (
		e           *echo.Echo
		recorder    *httptest.ResponseRecorder
		authService *handlerfakes.FakeService
		authHandler *handler.Handler
	)

	BeforeEach(func() {
		e = echo.New()
		recorder = httptest.NewRecorder()
		authService = &handlerfakes.FakeService{}

		authHandler = handler.NewHandler(authService)
		authHandler.RegisterHandlers(e)
	})

	Describe("POST /auth", func() {
		var (
			credentials model.AuthRequest
			token       string
			req         *http.Request
		)

		BeforeEach(func() {
			credentials = model.AuthRequest{
				Username: "testuser",
				Password: "testpass",
			}
			token = "testtoken"

			authService.AuthenticateReturns(token, nil)

			body, _ := json.Marshal(credentials)
			req = httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		})

		JustBeforeEach(func() {
			e.ServeHTTP(recorder, req)
		})

		It("succeeds", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
			Expect(authService.AuthenticateCallCount()).To(Equal(1))

			actualCtx, actualUsername, actualPassword := authService.AuthenticateArgsForCall(0)
			Expect(actualCtx).NotTo(BeNil())
			Expect(actualUsername).To(Equal(credentials.Username))
			Expect(actualPassword).To(Equal(credentials.Password))

			var response model.AuthResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Token).To(Equal(token))
		})

		When("the request contains invalid JSON", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodPost, "/auth", bytes.NewBufferString("{invalid json"))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			})

			It("returns 400", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("invalid request body"))
			})
		})

		When("the username or password is missing", func() {
			When("the username is missing", func() {
				BeforeEach(func() {
					credentials = model.AuthRequest{
						Username: "",
						Password: "testpass",
					}
					body, _ := json.Marshal(credentials)
					req = httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(body))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				})

				It("returns 400", func() {
					Expect(recorder.Code).To(Equal(http.StatusBadRequest))

					var response map[string]string
					err := json.Unmarshal(recorder.Body.Bytes(), &response)
					Expect(err).NotTo(HaveOccurred())
					Expect(response["message"]).To(Equal("username and password are required"))
				})
			})

			When("the password is missing", func() {
				BeforeEach(func() {
					credentials = model.AuthRequest{
						Username: "testuser",
						Password: "",
					}
					body, _ := json.Marshal(credentials)
					req = httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(body))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				})

				It("returns 400", func() {
					Expect(recorder.Code).To(Equal(http.StatusBadRequest))

					var response map[string]string
					err := json.Unmarshal(recorder.Body.Bytes(), &response)
					Expect(err).NotTo(HaveOccurred())
					Expect(response["message"]).To(Equal("username and password are required"))
				})
			})
		})

		When("authentication fails due to invalid credentials", func() {
			BeforeEach(func() {
				authService.AuthenticateReturns("", service.ErrInvalidCredentials)
			})

			It("returns 401", func() {
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("invalid credentials"))
			})
		})

		When("authentication fails due to an unexpected error", func() {
			BeforeEach(func() {
				authService.AuthenticateReturns("", ErrUnexpectedError)
			})

			It("returns 500", func() {
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("failed to authenticate user"))
			})
		})
	})
})
