package featureflags_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	authModel "github.com/georgisomnoev/feature-flag-api/internal/auth/model"
	authStore "github.com/georgisomnoev/feature-flag-api/internal/auth/store"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/handler"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/model"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/service"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/store"
	"github.com/georgisomnoev/feature-flag-api/internal/jwthelper"
	"github.com/georgisomnoev/feature-flag-api/internal/validator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature Flags Integration Test", func() {
	var (
		token               string
		userID              uuid.UUID
		srv                 *httptest.Server
		authenticationStore *authStore.Store
		featureFlagStore    *store.Store
		featureFlagService  *service.Service
		featureFlagHandler  *handler.Handler

		jwtPrivateKey = "../../certs/jwt_keys/private.pem"
		jwtPublicKey  = "../../certs/jwt_keys/public.pem"
	)

	BeforeEach(func() {
		e := echo.New()
		e.Validator = validator.GetValidator()
		jwtHelper, err := jwthelper.NewJWTHelper(jwtPrivateKey, jwtPublicKey)
		Expect(err).ToNot(HaveOccurred())

		featureFlagStore = store.NewStore(pool)
		featureFlagService = service.NewService(featureFlagStore)

		authenticationStore = authStore.NewStore(pool)
		featureFlagHandler = handler.NewHandler(featureFlagService, authenticationStore, jwtHelper)
		featureFlagHandler.RegisterHandlers(e)

		srv = httptest.NewServer(e)

		userID = uuid.New()
		claims := jwt.MapClaims{
			"sub":    userID,
			"scopes": []string{"read:flags", "write:flags"},
			"exp":    time.Now().Add(1 * time.Hour).Unix(),
		}
		token, err = jwtHelper.GenerateToken(claims)
		Expect(err).NotTo(HaveOccurred())

		user := authModel.User{ID: userID, Role: authModel.RoleEditor}
		err = authenticationStore.AddUser(ctx, user)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := authenticationStore.DeleteUserByID(ctx, userID)
		Expect(err).NotTo(HaveOccurred())
		srv.Close()
	})

	Describe("Feature Flags API", func() {
		var (
			testFlag model.FeatureFlag
		)

		BeforeEach(func() {
			testFlag = model.FeatureFlag{
				ID:          uuid.New(),
				Key:         "test-flag",
				Description: "test description",
				Enabled:     true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			err := featureFlagStore.CreateFlag(ctx, testFlag)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := featureFlagStore.DeleteFlag(ctx, testFlag.ID)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("List Feature Flags", func() {
			It("returns all feature flags", func() {
				req, err := http.NewRequest(http.MethodGet, srv.URL+"/flags", nil)
				Expect(err).ToNot(HaveOccurred())
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

				resp, err := http.DefaultClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				var flags []model.FeatureFlag
				err = json.NewDecoder(resp.Body).Decode(&flags)
				Expect(err).ToNot(HaveOccurred())
				Expect(flags).To(HaveLen(1))
				Expect(flags[0].Key).To(Equal("test-flag"))
			})
		})

		Context("Get Feature Flag By ID", func() {
			It("returns the feature flag", func() {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/flags/%s", srv.URL, testFlag.ID), nil)
				Expect(err).ToNot(HaveOccurred())
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

				resp, err := http.DefaultClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				var flag model.FeatureFlag
				err = json.NewDecoder(resp.Body).Decode(&flag)
				Expect(err).ToNot(HaveOccurred())
				Expect(flag.Key).To(Equal("test-flag"))
			})
		})

		Context("Create New Feature Flag", func() {
			var (
				payload []byte
				req     *http.Request
				resp    *http.Response
				err     error
			)

			BeforeEach(func() {
				payload, err = json.Marshal(map[string]interface{}{
					"key":         "new-flag",
					"description": "new description",
					"enabled":     true,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			JustBeforeEach(func() {
				req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/flags", srv.URL), bytes.NewBuffer(payload))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				req.Header.Set("Content-Type", "application/json")
				resp, err = http.DefaultClient.Do(req)
				Expect(err).NotTo(HaveOccurred())
			})

			JustAfterEach(func() {
				// TODO: adjust to proper cleanup.
				insertedFlag, err := featureFlagStore.GetFlagByKey(ctx, "new-flag")
				Expect(err).ToNot(HaveOccurred())
				err = featureFlagStore.DeleteFlag(ctx, insertedFlag.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("creates the feature flag", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			})
		})

		Context("Update Existing Feature Flag", func() {
			It("updates the feature flag", func() {
				payload, _ := json.Marshal(map[string]interface{}{
					"key":         "updated-flag",
					"description": "updated description",
					"enabled":     false,
				})

				req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/flags/%s", srv.URL, testFlag.ID), bytes.NewBuffer(payload))
				Expect(err).ToNot(HaveOccurred())
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("Delete Feature Flag", func() {
			It("deletes the feature flag", func() {
				req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/flags/%s", srv.URL, testFlag.ID), nil)
				Expect(err).ToNot(HaveOccurred())
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

				resp, err := http.DefaultClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
			})
		})
	})
})
