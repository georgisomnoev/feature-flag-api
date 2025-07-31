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
	. "github.com/onsi/gomega/gstruct"
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

		flagID uuid.UUID
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
			Expect(flags).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"ID":          Equal(featureFlag.ID),
				"Key":         Equal(featureFlag.Key),
				"Description": Equal(featureFlag.Description),
				"Enabled":     Equal(featureFlag.Enabled),
			})))
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
			Expect(store.GetFlagByIDCallCount()).To(Equal(1))
			Expect(featureFlag).To((MatchFields(IgnoreExtras, Fields{
				"ID":          Equal(flagID),
				"Key":         Equal(featureFlag.Key),
				"Description": Equal(featureFlag.Description),
				"Enabled":     Equal(featureFlag.Enabled),
			})))
		})

		Context("when the store returns not found", func() {
			BeforeEach(func() {
				store.GetFlagByIDReturns(model.FeatureFlag{}, model.ErrNotFound)
			})

			It("returns the not found error", func() {
				Expect(errAction).To(MatchError(model.ErrNotFound))
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
			featureFlagRequest model.FeatureFlagRequest
		)

		BeforeEach(func() {
			featureFlagRequest = model.FeatureFlagRequest{Key: "new-flag", Description: "new description", Enabled: true}
		})

		JustBeforeEach(func() {
			flagID, errAction = svc.CreateFlag(ctx, featureFlagRequest)
		})

		ItSucceeds()
		It("creates the feature flag", func() {
			Expect(store.CreateFlagCallCount()).To(Equal(1))
			actualCtx, actualFlag := store.CreateFlagArgsForCall(0)
			Expect(actualCtx).To(Equal(ctx))
			Expect(actualFlag).To((MatchFields(IgnoreExtras, Fields{
				"ID":          Equal(flagID),
				"Key":         Equal(featureFlagRequest.Key),
				"Description": Equal(featureFlagRequest.Description),
				"Enabled":     Equal(featureFlagRequest.Enabled),
			})))
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
			featureFlagRequest model.FeatureFlagRequest
			newUUID            uuid.UUID
		)

		BeforeEach(func() {
			featureFlagRequest = model.FeatureFlagRequest{Key: "updated-flag", Description: "updated description", Enabled: false}
			store.UpdateFlagReturns(nil)
		})

		JustBeforeEach(func() {
			errAction = svc.UpdateFlag(ctx, newUUID, featureFlagRequest)
		})

		ItSucceeds()
		It("updates the feature flag", func() {
			Expect(store.UpdateFlagCallCount()).To(Equal(1))
			actualCtx, actualFlag := store.UpdateFlagArgsForCall(0)
			Expect(actualCtx).To(Equal(ctx))
			Expect(actualFlag.ID).To(Equal(newUUID))
			Expect(actualFlag.Key).To(Equal(featureFlagRequest.Key))
			Expect(actualFlag.Description).To(Equal(featureFlagRequest.Description))
			Expect(actualFlag.Enabled).To(BeFalse())
		})

		Context("when the flag does not exist", func() {
			BeforeEach(func() {
				store.UpdateFlagReturns(model.ErrNotFound)
			})

			It("returns the not found error", func() {
				Expect(errAction).To(MatchError(model.ErrNotFound))
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
				store.DeleteFlagReturns(model.ErrNotFound)
			})

			It("returns the not found error", func() {
				Expect(errAction).To(MatchError(model.ErrNotFound))
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
