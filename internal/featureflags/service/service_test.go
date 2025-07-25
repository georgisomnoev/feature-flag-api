package service_test

import (
	"context"
	"errors"

	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/model"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/service"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/service/servicefakes"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	ErrDatabaseError = errors.New("database error")
)

var _ = Describe("Service", func() {
	var (
		ctx       context.Context
		errAction error
		svc       *service.Service
		store     *servicefakes.FakeStore
	)

	BeforeEach(func() {
		ctx = context.Background()
		store = &servicefakes.FakeStore{}
		svc = service.NewService(store)
	})

	ItSucceeds := func() {
		It("succeeds", func() {
			Expect(errAction).ToNot(HaveOccurred())
		})
	}

	Describe("ListFlags", func() {
		var (
			flags       []model.FeatureFlag
			featureFlag model.FeatureFlag
		)

		BeforeEach(func() {
			featureFlag = model.FeatureFlag{ID: uuid.New(), Key: "test-flag", Description: "description", Enabled: true}
			store.ListFlagsReturns([]model.FeatureFlag{featureFlag}, nil)
		})

		JustBeforeEach(func() {
			flags, errAction = svc.ListFlags(ctx)
		})

		ItSucceeds()
		It("returns the list of feature flags", func() {
			Expect(flags).To(HaveLen(1))
			Expect(flags[0].Key).To(Equal(featureFlag.Key))
			Expect(flags[0].Description).To(Equal(featureFlag.Description))
			Expect(flags[0].Enabled).To(Equal(featureFlag.Enabled))
		})

		Context("when the store returns an error", func() {
			BeforeEach(func() {
				store.ListFlagsReturns(nil, ErrDatabaseError)
			})

			It("returns the error", func() {
				Expect(errAction).To(MatchError(ErrDatabaseError))
			})
		})
	})

	Describe("GetFlagByID", func() {
		var (
			featureFlag model.FeatureFlag
			flagID      uuid.UUID
		)

		BeforeEach(func() {
			flagID = uuid.New()
			featureFlag = model.FeatureFlag{ID: flagID, Key: "test-flag", Description: "test description", Enabled: true}
			store.GetFlagByIDReturns(featureFlag, nil)
		})

		JustBeforeEach(func() {
			featureFlag, errAction = svc.GetFlagByID(ctx, flagID)
		})

		ItSucceeds()
		It("returns the feature flag by ID", func() {
			Expect(featureFlag.Key).To(Equal("test-flag"))
			Expect(featureFlag.Description).To(Equal("test description"))
			Expect(featureFlag.Enabled).To(BeTrue())
		})

		Context("when the store returns not found", func() {
			BeforeEach(func() {
				store.GetFlagByIDReturns(model.FeatureFlag{}, service.ErrNotFound)
			})

			It("returns the not found error", func() {
				Expect(errAction).To(MatchError(service.ErrNotFound))
			})
		})

		Context("when the store returns another error", func() {
			BeforeEach(func() {
				store.GetFlagByIDReturns(model.FeatureFlag{}, ErrDatabaseError)
			})

			It("returns the error", func() {
				Expect(errAction).To(MatchError(ErrDatabaseError))
			})
		})
	})

	Describe("CreateFlag", func() {
		var (
			featureFlag model.FeatureFlag
		)

		BeforeEach(func() {
			featureFlag = model.FeatureFlag{Key: "new-flag", Description: "new description", Enabled: true}
			store.CreateFlagReturns(nil)
		})

		JustBeforeEach(func() {
			errAction = svc.CreateFlag(ctx, featureFlag)
		})

		ItSucceeds()
		It("creates the feature flag", func() {
			Expect(store.CreateFlagCallCount()).To(Equal(1))
			actualCtx, actualFlag := store.CreateFlagArgsForCall(0)
			Expect(actualCtx).To(Equal(ctx))
			Expect(actualFlag.Key).To(Equal(featureFlag.Key))
			Expect(actualFlag.Description).To(Equal(featureFlag.Description))
		})

		Context("when the store returns an error", func() {
			BeforeEach(func() {
				store.CreateFlagReturns(ErrDatabaseError)
			})

			It("returns the error", func() {
				Expect(errAction).To(MatchError(ErrDatabaseError))
			})
		})
	})

	Describe("UpdateFlag", func() {
		var (
			featureFlag model.FeatureFlag
		)

		BeforeEach(func() {
			featureFlag = model.FeatureFlag{ID: uuid.New(), Key: "updated-flag", Description: "updated description", Enabled: false}
			store.UpdateFlagReturns(nil)
		})

		JustBeforeEach(func() {
			errAction = svc.UpdateFlag(ctx, featureFlag)
		})

		ItSucceeds()
		It("updates the feature flag", func() {
			Expect(store.UpdateFlagCallCount()).To(Equal(1))
			actualCtx, actualFlag := store.UpdateFlagArgsForCall(0)
			Expect(actualCtx).To(Equal(ctx))
			Expect(actualFlag.Key).To(Equal(featureFlag.Key))
			Expect(actualFlag.Description).To(Equal(featureFlag.Description))
		})

		Context("when the flag does not exist", func() {
			BeforeEach(func() {
				store.UpdateFlagReturns(service.ErrNotFound)
			})

			It("returns the not found error", func() {
				Expect(errAction).To(MatchError(service.ErrNotFound))
			})
		})

		Context("when the store returns another error", func() {
			BeforeEach(func() {
				store.UpdateFlagReturns(ErrDatabaseError)
			})

			It("returns the error", func() {
				Expect(errAction).To(MatchError(ErrDatabaseError))
			})
		})
	})

	Describe("DeleteFlag", func() {
		var flagID uuid.UUID

		BeforeEach(func() {
			flagID = uuid.New()
			store.DeleteFlagReturns(nil)
		})

		JustBeforeEach(func() {
			errAction = svc.DeleteFlag(ctx, flagID)
		})

		ItSucceeds()
		It("deletes the feature flag", func() {
			Expect(store.DeleteFlagCallCount()).To(Equal(1))
			actualCtx, actualFlagID := store.DeleteFlagArgsForCall(0)
			Expect(actualCtx).To(Equal(ctx))
			Expect(actualFlagID).To(Equal(flagID))
		})

		Context("when the flag does not exist", func() {
			BeforeEach(func() {
				store.DeleteFlagReturns(service.ErrNotFound)
			})

			It("returns the not found error", func() {
				Expect(errAction).To(MatchError(service.ErrNotFound))
			})
		})

		Context("when the store returns another error", func() {
			BeforeEach(func() {
				store.DeleteFlagReturns(ErrDatabaseError)
			})

			It("returns the error", func() {
				Expect(errAction).To(MatchError(ErrDatabaseError))
			})
		})
	})
})
