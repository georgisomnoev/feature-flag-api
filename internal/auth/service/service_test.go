package service_test

import (
	"context"
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"

	"github.com/georgisomnoev/feature-flag-api/internal/auth/model"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/service"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/service/servicefakes"
)

var (
	ErrDatabaseError         = errors.New("database error")
	ErrTokenGenerationFailed = errors.New("failed to generate token")
)

var _ = Describe("Service", func() {
	var (
		store     *servicefakes.FakeStore
		jwtHelper *servicefakes.FakeJWTHelper
		svc       *service.Service
	)
	BeforeEach(func() {
		store = &servicefakes.FakeStore{}
		jwtHelper = &servicefakes.FakeJWTHelper{}
		svc = service.NewService(store, jwtHelper)
	})
	Describe("Authenticate", func() {
		var (
			token     string
			user      model.User
			ctx       context.Context
			errAction error

			username string
			password string
		)

		BeforeEach(func() {
			username = "testuser"
			password = "testpass"
			ctx = context.Background()

			// The default cost for bcrypt is 10, which is slow for the tests.
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			Expect(err).ToNot(HaveOccurred())

			user = model.User{
				ID:       uuid.New(),
				Username: username,
				Password: string(hashedPassword),
				Role:     model.RoleEditor,
			}

			store.GetByUsernameReturns(&user, nil)
		})

		ItSucceeds := func() {
			It("succeeds", func() {
				Expect(errAction).ToNot(HaveOccurred())
			})
		}

		JustBeforeEach(func() {
			token, errAction = svc.Authenticate(ctx, username, password)
		})

		Context("when authentication succeeds for RoleEditor", func() {
			BeforeEach(func() {
				jwtHelper.GenerateTokenReturns("valid-editor-token", nil)
			})

			ItSucceeds()
			It("returns a token with editor scopes", func() {
				Expect(token).To(Equal("valid-editor-token"))

				claims := jwtHelper.GenerateTokenArgsForCall(0)
				Expect(claims).To(HaveKeyWithValue("sub", user.ID))
				Expect(claims).To(HaveKeyWithValue("scopes", []string{"read:flags", "write:flags"}))
			})
		})

		Context("when authentication succeeds for RoleViewer", func() {
			BeforeEach(func() {
				user.Role = model.RoleViewer
				store.GetByUsernameReturns(&user, nil)
				jwtHelper.GenerateTokenReturns("valid-viewer-token", nil)
			})

			ItSucceeds()
			It("returns a token with viewer scopes", func() {
				Expect(token).To(Equal("valid-viewer-token"))

				claims := jwtHelper.GenerateTokenArgsForCall(0)
				Expect(claims).To(HaveKeyWithValue("sub", user.ID))
				Expect(claims).To(HaveKeyWithValue("scopes", []string{"read:flags"}))
			})
		})

		Context("when invalid credentials are provided", func() {
			BeforeEach(func() {
				password = "wrongpassword"
			})

			It("returns an invalid credentials error", func() {
				Expect(errAction).To(MatchError(service.ErrInvalidCredentials))
				Expect(token).To(BeEmpty())
			})
		})

		Context("when the user role is invalid", func() {
			BeforeEach(func() {
				user.Role = "invalidRole"
				store.GetByUsernameReturns(&user, nil)
			})

			It("returns an invalid user role error", func() {
				Expect(errAction).To(MatchError(service.ErrInvalidUserRole))
				Expect(token).To(BeEmpty())
			})
		})

		Context("when the token generation fails", func() {
			BeforeEach(func() {
				jwtHelper.GenerateTokenReturns("", ErrTokenGenerationFailed)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrTokenGenerationFailed))
				Expect(token).To(BeEmpty())
			})
		})

		Context("when user is not found", func() {
			BeforeEach(func() {
				store.GetByUsernameReturns(nil, nil)
			})

			It("returns an invalid credentials error", func() {
				Expect(errAction).To(MatchError(service.ErrInvalidCredentials))
				Expect(token).To(BeEmpty())
			})
		})

		Context("when there is an error retrieving the user", func() {
			BeforeEach(func() {
				store.GetByUsernameReturns(nil, ErrDatabaseError)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrDatabaseError))
				Expect(token).To(BeEmpty())
			})
		})
	})
})
