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

var _ = Describe("Feature Flags Integration Test", Label("integration"), func() {
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
			testFlag   model.FeatureFlag
			testFlagID uuid.UUID
			errAction  error

			req  *http.Request
			resp *http.Response
		)

		BeforeEach(func() {
			testFlagID = uuid.New()
			testFlag = model.FeatureFlag{
				ID:          testFlagID,
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

		JustBeforeEach(func() {
			resp, errAction = http.DefaultClient.Do(req)

			DeferCleanup(func() {
				if resp != nil {
					resp.Body.Close()
				}
			})
		})

		ItSucceeds := func() {
			It("succeeds", func() {
				Expect(errAction).ToNot(HaveOccurred())
			})
		}

		Context("List Feature Flags", func() {
			var (
				err error
			)

			BeforeEach(func() {
				req, err = http.NewRequest(http.MethodGet, srv.URL+"/flags", nil)
				Expect(err).ToNot(HaveOccurred())
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			})

			ItSucceeds()
			It("returns the feature flag previously added", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				var flags []model.FeatureFlag
				err = json.NewDecoder(resp.Body).Decode(&flags)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(flags)).To(BeNumerically(">=", 1))
				Expect(flags[len(flags)-1].ID).To(Equal(testFlag.ID))
				Expect(flags[len(flags)-1].Key).To(Equal(testFlag.Key))
				Expect(flags[len(flags)-1].Description).To(Equal(testFlag.Description))
				Expect(flags[len(flags)-1].CreatedAt).To(BeTemporally(">", time.Now()))
				Expect(flags[len(flags)-1].UpdatedAt).To(BeTemporally(">", time.Now()))

			})
		})

		Context("Get Feature Flag By ID", func() {
			var (
				err error
			)

			BeforeEach(func() {
				req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/flags/%s", srv.URL, testFlag.ID), nil)
				Expect(err).ToNot(HaveOccurred())
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			})

			ItSucceeds()
			It("returns the feature flag", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				var flag model.FeatureFlag
				err = json.NewDecoder(resp.Body).Decode(&flag)
				Expect(err).ToNot(HaveOccurred())
				Expect(flag.ID).To(Equal(testFlagID))
				Expect(flag.Key).To(Equal(testFlag.Key))
				Expect(flag.Description).To(Equal(testFlag.Description))
				Expect(flag.CreatedAt).To(BeTemporally(">", time.Now()))
				Expect(flag.UpdatedAt).To(BeTemporally(">", time.Now()))
			})
		})

		Context("Create New Feature Flag", func() {
			var (
				generateFlagID uuid.UUID
			)

			BeforeEach(func() {
				payload, err := json.Marshal(map[string]interface{}{
					"key":         "new-flag",
					"description": "new description",
					"enabled":     true,
				})
				Expect(err).NotTo(HaveOccurred())

				req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/flags", srv.URL), bytes.NewBuffer(payload))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				req.Header.Set("Content-Type", "application/json")
			})

			JustBeforeEach(func() {
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				var response struct {
					ID uuid.UUID `json:"id"`
				}
				var err = json.NewDecoder(resp.Body).Decode(&response)
				Expect(err).NotTo(HaveOccurred())
				generateFlagID = response.ID
			})

			JustAfterEach(func() {
				err := featureFlagStore.DeleteFlag(ctx, generateFlagID)
				Expect(err).ToNot(HaveOccurred())
			})

			ItSucceeds()
			It("creates the feature flag", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			})
		})

		Context("Update Existing Feature Flag", func() {
			var (
				payload     []byte
				anotherFlag model.FeatureFlag
			)

			BeforeEach(func() {
				anotherFlag = model.FeatureFlag{
					ID:          uuid.New(),
					Key:         testFlag.Key,
					Description: testFlag.Description,
					Enabled:     true,
				}

				err := featureFlagStore.CreateFlag(ctx, anotherFlag)
				Expect(err).ToNot(HaveOccurred())

				payload, err = json.Marshal(map[string]interface{}{
					"key":         "updated-flag",
					"description": "updated description",
					"enabled":     false,
				})
				Expect(err).ToNot(HaveOccurred())

				req, err = http.NewRequest(http.MethodPut, fmt.Sprintf("%s/flags/%s", srv.URL, anotherFlag.ID), bytes.NewBuffer(payload))
				Expect(err).ToNot(HaveOccurred())
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				req.Header.Set("Content-Type", "application/json")
			})

			AfterEach(func() {
				err := featureFlagStore.DeleteFlag(ctx, anotherFlag.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			ItSucceeds()
			It("updates the feature flag", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("Delete Feature Flag", func() {
			var (
				anotherFlag model.FeatureFlag
			)

			BeforeEach(func() {
				anotherFlag = model.FeatureFlag{
					ID:          uuid.New(),
					Key:         testFlag.Key,
					Description: testFlag.Description,
					Enabled:     true,
				}

				err := featureFlagStore.CreateFlag(ctx, anotherFlag)
				Expect(err).ToNot(HaveOccurred())

				req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/flags/%s", srv.URL, anotherFlag.ID), nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			})

			ItSucceeds()
			It("deletes the feature flag", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
			})
		})
	})
})
