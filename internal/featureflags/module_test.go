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
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/model"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/store"
	"github.com/georgisomnoev/feature-flag-api/internal/jwthelper"
	"github.com/georgisomnoev/feature-flag-api/internal/validator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Feature Flags Integration Test", Label("integration"), func() {
	var (
		token               string
		userID              uuid.UUID
		srv                 *httptest.Server
		authenticationStore *authStore.Store
		featureFlagStore    *store.Store

		jwtPrivateKey = "../../certs/jwt_keys/private.pem"
		jwtPublicKey  = "../../certs/jwt_keys/public.pem"
	)

	BeforeEach(func() {
		e := echo.New()
		e.Validator = validator.GetValidator()
		authenticationStore = authStore.NewStore(pool)
		jwtHelper, err := jwthelper.NewJWTHelper(jwtPrivateKey, jwtPublicKey)
		Expect(err).ToNot(HaveOccurred())

		featureFlagStore = store.NewStore(pool)

		featureflags.Process(pool, e, authenticationStore, jwtHelper)

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
			testFlag  model.FeatureFlag
			errAction error

			req  *http.Request
			resp *http.Response
		)

		BeforeEach(func() {
			testFlag = model.FeatureFlag{
				ID:          uuid.New(),
				Key:         "test-flag",
				Description: "test description",
				Enabled:     true,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
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
				req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
			})

			ItSucceeds()
			It("returns the feature flag previously added", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				var flags []model.FeatureFlag
				err = json.NewDecoder(resp.Body).Decode(&flags)
				Expect(err).NotTo(HaveOccurred())
				Expect(flags).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"ID":          Equal(testFlag.ID),
					"Key":         Equal(testFlag.Key),
					"Description": Equal(testFlag.Description),
					"Enabled":     Equal(testFlag.Enabled),
					"CreatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
					"UpdatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
				})))
			})
		})

		Context("Get Feature Flag By ID", func() {
			var (
				err error
			)

			BeforeEach(func() {
				req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/flags/%s", srv.URL, testFlag.ID), nil)
				Expect(err).ToNot(HaveOccurred())
				req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
			})

			ItSucceeds()
			It("returns the feature flag", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				var flag model.FeatureFlag
				err = json.NewDecoder(resp.Body).Decode(&flag)
				Expect(err).ToNot(HaveOccurred())
				Expect(flag).To((MatchFields(IgnoreExtras, Fields{
					"ID":          Equal(testFlag.ID),
					"Key":         Equal(testFlag.Key),
					"Description": Equal(testFlag.Description),
					"Enabled":     Equal(testFlag.Enabled),
					"CreatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
					"UpdatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
				})))
			})
		})

		Context("Create New Feature Flag", func() {
			var (
				generateFlagID uuid.UUID
				newFlag        model.FeatureFlag
			)

			BeforeEach(func() {
				newFlag = model.FeatureFlag{
					Key:         "new-flag",
					Description: "new description",
					Enabled:     true,
				}
				payload, err := json.Marshal(newFlag)
				Expect(err).NotTo(HaveOccurred())

				req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/flags", srv.URL), bytes.NewBuffer(payload))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
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
			It("returns the correct status code", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			})
			It("creates the flag", func() {
				storedFlag, err := featureFlagStore.GetFlagByID(ctx, generateFlagID)
				Expect(err).NotTo(HaveOccurred())
				Expect(storedFlag).To((MatchFields(IgnoreExtras, Fields{
					"ID":          Equal(generateFlagID),
					"Key":         Equal(newFlag.Key),
					"Description": Equal(newFlag.Description),
					"Enabled":     Equal(newFlag.Enabled),
					"CreatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
					"UpdatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
				})))
			})
		})

		Context("Update Existing Feature Flag", func() {
			var (
				payload     []byte
				anotherFlag model.FeatureFlag
				updatedFlag model.FeatureFlag
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

				updatedFlag = model.FeatureFlag{
					Key:         "updated-flag",
					Description: "updated description",
					Enabled:     false,
				}
				payload, err = json.Marshal(updatedFlag)
				Expect(err).ToNot(HaveOccurred())

				req, err = http.NewRequest(http.MethodPut, fmt.Sprintf("%s/flags/%s", srv.URL, anotherFlag.ID), bytes.NewBuffer(payload))
				Expect(err).ToNot(HaveOccurred())
				req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			})

			AfterEach(func() {
				err := featureFlagStore.DeleteFlag(ctx, anotherFlag.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			ItSucceeds()
			It("returns the correct status code", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
			It("updates the flag", func() {
				storedFlag, err := featureFlagStore.GetFlagByID(ctx, anotherFlag.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(storedFlag).To((MatchFields(IgnoreExtras, Fields{
					"ID":          Equal(anotherFlag.ID),
					"Key":         Equal(updatedFlag.Key),
					"Description": Equal(updatedFlag.Description),
					"Enabled":     Equal(updatedFlag.Enabled),
					"CreatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
					"UpdatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
				})))
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
				req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
			})

			ItSucceeds()
			It("returns the correct status code", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
			})
			It("deletes the flag", func() {
				_, err := featureFlagStore.GetFlagByID(ctx, anotherFlag.ID)
				Expect(err).To(MatchError(model.ErrNotFound))
			})
		})
	})
})
