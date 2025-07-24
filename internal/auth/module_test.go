package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/georgisomnoev/feature-flag-api/internal/auth"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/model"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/store"
	"github.com/georgisomnoev/feature-flag-api/internal/jwthelper"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Integration Test", func() {
	var (
		srv       *httptest.Server
		authStore *store.Store

		jwtPrivateKey = "../../certs/jwt_keys/private.pem"
		jwtPublicKey  = "../../certs/jwt_keys/public.pem"
	)

	BeforeEach(func() {
		e := echo.New()
		jwtHelper, err := jwthelper.NewJWTHelper(jwtPrivateKey, jwtPublicKey)
		Expect(err).ToNot(HaveOccurred())

		authStore = auth.Process(ctx, pool, e, jwtHelper)
		Expect(authStore).ToNot(BeNil())

		srv = httptest.NewServer(e)
	})

	AfterEach(func() {
		srv.Close()
	})

	Describe("Authenticate API", func() {
		var (
			username string
			password string
			testUser model.User
		)

		BeforeEach(func() {
			username = "testuser"
			password = "testpass"
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			Expect(err).ToNot(HaveOccurred())

			testUser = model.User{
				ID:       uuid.New(),
				Username: username,
				Password: string(hashedPassword),
				Role:     "editor",
			}

			err = authStore.AddUser(ctx, testUser)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := authStore.DeleteUserByID(ctx, testUser.ID)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when authenticating an existing user", func() {
			It("returns a valid token", func() {
				payload, _ := json.Marshal(map[string]string{
					"username": testUser.Username,
					"password": "testpass",
				})

				resp, err := http.Post(srv.URL+"/auth", "application/json", bytes.NewBuffer(payload))
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				var response map[string]string
				err = json.NewDecoder(resp.Body).Decode(&response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("token"))
				Expect(response["token"]).ToNot(BeEmpty())
			})
		})

		Context("when authentication fails due to incorrect password", func() {
			It("returns an unauthorized error", func() {
				payload, _ := json.Marshal(map[string]string{
					"username": testUser.Username,
					"password": "wrongpass",
				})

				resp, err := http.Post(srv.URL+"/auth", "application/json", bytes.NewBuffer(payload))
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))

				var response map[string]string
				err = json.NewDecoder(resp.Body).Decode(&response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("message"))
				Expect(response["message"]).To(Equal("invalid credentials"))
			})
		})

		Context("when the user does not exist", func() {
			It("returns an unauthorized error", func() {
				payload, _ := json.Marshal(map[string]string{
					"username": "nonexistentuser",
					"password": "testpass",
				})

				resp, err := http.Post(srv.URL+"/auth", "application/json", bytes.NewBuffer(payload))
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))

				var response map[string]string
				err = json.NewDecoder(resp.Body).Decode(&response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("message"))
				Expect(response["message"]).To(Equal("invalid credentials"))
			})
		})
	})
})
