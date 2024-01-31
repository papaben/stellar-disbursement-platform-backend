package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/stellar/go/support/log"
	"github.com/stellar/stellar-disbursement-platform-backend/db"
	"github.com/stellar/stellar-disbursement-platform-backend/internal/data"
	"github.com/stellar/stellar-disbursement-platform-backend/internal/events"
	"github.com/stellar/stellar-disbursement-platform-backend/internal/events/schemas"
	"github.com/stellar/stellar-disbursement-platform-backend/internal/utils"
	"github.com/stellar/stellar-disbursement-platform-backend/stellar-auth/pkg/auth"
)

// DisbursementManagementService is a service for managing disbursements.
type DisbursementManagementService struct {
	models           *data.Models
	dbConnectionPool db.DBConnectionPool
	eventProducer    events.Producer
}

var (
	ErrDisbursementNotFound        = errors.New("disbursement not found")
	ErrDisbursementNotReadyToStart = errors.New("disbursement is not ready to be started")
	ErrDisbursementNotReadyToPause = errors.New("disbursement is not ready to be paused")
	ErrDisbursementWalletDisabled  = errors.New("disbursement wallet is disabled")

	ErrDisbursementStatusCantBeChanged = errors.New("disbursement status can't be changed to the requested status")
	ErrDisbursementStartedByCreator    = errors.New("disbursement can't be started by its creator")
)

// NewDisbursementManagementService is a factory function for creating a new DisbursementManagementService.
func NewDisbursementManagementService(models *data.Models, dbConnectionPool db.DBConnectionPool, eventProducer events.Producer) *DisbursementManagementService {
	return &DisbursementManagementService{
		models:           models,
		dbConnectionPool: dbConnectionPool,
		eventProducer:    eventProducer,
	}
}

func (s *DisbursementManagementService) GetDisbursementsWithCount(ctx context.Context, queryParams *data.QueryParams) (*utils.ResultWithTotal, error) {
	return db.RunInTransactionWithResult(ctx,
		s.dbConnectionPool,
		&sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true},
		func(dbTx db.DBTransaction) (*utils.ResultWithTotal, error) {
			totalDisbursements, err := s.models.Disbursements.Count(ctx, dbTx, queryParams)
			if err != nil {
				return nil, fmt.Errorf("error counting disbursements: %w", err)
			}

			var disbursements []*data.Disbursement
			if totalDisbursements != 0 {
				disbursements, err = s.models.Disbursements.GetAll(ctx, dbTx, queryParams)
				if err != nil {
					return nil, fmt.Errorf("error retrieving disbursements: %w", err)
				}
			}

			return utils.NewResultWithTotal(totalDisbursements, disbursements), nil
		})
}

func (s *DisbursementManagementService) GetDisbursementReceiversWithCount(ctx context.Context, disbursementID string, queryParams *data.QueryParams) (*utils.ResultWithTotal, error) {
	return db.RunInTransactionWithResult(ctx,
		s.dbConnectionPool,
		&sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true},
		func(dbTx db.DBTransaction) (*utils.ResultWithTotal, error) {
			_, err := s.models.Disbursements.Get(ctx, dbTx, disbursementID)
			if err != nil {
				if errors.Is(err, data.ErrRecordNotFound) {
					return nil, ErrDisbursementNotFound
				} else {
					return nil, fmt.Errorf("error getting disbursement with id %s: %w", disbursementID, err)
				}
			}

			totalReceivers, err := s.models.DisbursementReceivers.Count(ctx, dbTx, disbursementID)
			if err != nil {
				return nil, fmt.Errorf("error counting disbursement receivers for disbursement with id %s: %w", disbursementID, err)
			}

			receivers := []*data.DisbursementReceiver{}
			if totalReceivers != 0 {
				receivers, err = s.models.DisbursementReceivers.GetAll(ctx, dbTx, queryParams, disbursementID)
				if err != nil {
					return nil, fmt.Errorf("error retrieving disbursement receivers for disbursement with id %s: %w", disbursementID, err)
				}
			}

			return utils.NewResultWithTotal(totalReceivers, receivers), nil
		})
}

// StartDisbursement starts a disbursement and all its payments and receivers wallets.
func (s *DisbursementManagementService) StartDisbursement(ctx context.Context, disbursementID string, user *auth.User) error {
	return db.RunInTransaction(ctx, s.dbConnectionPool, nil, func(dbTx db.DBTransaction) error {
		disbursement, err := s.models.Disbursements.Get(ctx, dbTx, disbursementID)
		if err != nil {
			if errors.Is(err, data.ErrRecordNotFound) {
				return ErrDisbursementNotFound
			} else {
				return fmt.Errorf("error getting disbursement with id %s: %w", disbursementID, err)
			}
		}

		// 1. Verify Wallet is Enabled
		if !disbursement.Wallet.Enabled {
			return ErrDisbursementWalletDisabled
		}
		// 2. Verify Transition is Possible
		err = disbursement.Status.TransitionTo(data.StartedDisbursementStatus)
		if err != nil {
			return ErrDisbursementNotReadyToStart
		}

		// 3. Check if approval Workflow is enabled for this organization
		organization, err := s.models.Organizations.Get(ctx)
		if err != nil {
			return fmt.Errorf("error getting organization: %w", err)
		}

		if organization.IsApprovalRequired {
			// check that the user starting the disbursement isn't the same as the one who created it
			for _, sh := range disbursement.StatusHistory {
				if sh.UserID == user.ID && (sh.Status == data.DraftDisbursementStatus || sh.Status == data.ReadyDisbursementStatus) {
					return ErrDisbursementStartedByCreator
				}
			}
		}

		// 4. Update all correct payment status to `ready`
		err = s.models.Payment.UpdateStatusByDisbursementID(ctx, dbTx, disbursementID, data.ReadyPaymentStatus)
		if err != nil {
			return fmt.Errorf("error updating payment status to ready for disbursement with id %s: %w", disbursementID, err)
		}

		// 5. Update all receiver_wallets from `draft` to `ready`
		err = s.models.ReceiverWallet.UpdateStatusByDisbursementID(ctx, dbTx, disbursementID, data.DraftReceiversWalletStatus, data.ReadyReceiversWalletStatus)
		if err != nil {
			return fmt.Errorf("error updating receiver wallet status to ready for disbursement with id %s: %w", disbursementID, err)
		}

		// 6. Update disbursement status to `started`
		err = s.models.Disbursements.UpdateStatus(ctx, dbTx, user.ID, disbursementID, data.StartedDisbursementStatus)
		if err != nil {
			return fmt.Errorf("error updating disbursement status to started for disbursement with id %s: %w", disbursementID, err)
		}

		// 7. Produce event to send payments to TSS
		payments, err := s.models.Payment.GetReadyByDisbursementID(ctx, dbTx, disbursementID)
		if err != nil {
			return fmt.Errorf("getting ready payments for disbursement with id %s: %w", disbursementID, err)
		}

		if len(payments) == 0 {
			log.Ctx(ctx).Infof("no payments ready to pay for disbursement ID %s", disbursementID)
			return nil
		}

		msg, err := events.NewMessage(ctx, events.PaymentReadyToPayTopic, disbursementID, events.PaymentReadyToPayDisbursementStarted, nil)
		if err != nil {
			return fmt.Errorf("creating new message: %w", err)
		}

		paymentsReadyToPay := schemas.EventPaymentsReadyToPayData{TenantID: msg.TenantID}
		for _, payment := range payments {
			paymentsReadyToPay.Payments = append(paymentsReadyToPay.Payments, schemas.PaymentReadyToPay{ID: payment.ID})
		}
		msg.Data = paymentsReadyToPay

		if s.eventProducer != nil {
			err := s.eventProducer.WriteMessages(ctx, *msg)
			if err != nil {
				return fmt.Errorf("writing message %s on event producer: %w", msg, err)
			}
		} else {
			log.Ctx(ctx).Errorf("event producer is nil, could not publish message %s", msg.String())
		}

		return nil
	})
}

// PauseDisbursement pauses a disbursement and all its payments.
func (s *DisbursementManagementService) PauseDisbursement(ctx context.Context, disbursementID string, user *auth.User) error {
	return db.RunInTransaction(ctx, s.dbConnectionPool, nil, func(dbTx db.DBTransaction) error {
		disbursement, err := s.models.Disbursements.Get(ctx, dbTx, disbursementID)
		if err != nil {
			if errors.Is(err, data.ErrRecordNotFound) {
				return ErrDisbursementNotFound
			} else {
				return fmt.Errorf("error getting disbursement with id %s: %w", disbursementID, err)
			}
		}

		// 1. Verify Transition is Possible
		err = disbursement.Status.TransitionTo(data.PausedDisbursementStatus)
		if err != nil {
			return ErrDisbursementNotReadyToPause
		}

		// 2. Update all correct payment status to `paused`
		err = s.models.Payment.UpdateStatusByDisbursementID(ctx, dbTx, disbursementID, data.PausedPaymentStatus)
		if err != nil {
			return fmt.Errorf("error updating payment status to paused for disbursement with id %s: %w", disbursementID, err)
		}

		// 3. Update disbursement status to `paused`
		err = s.models.Disbursements.UpdateStatus(ctx, dbTx, user.ID, disbursementID, data.PausedDisbursementStatus)
		if err != nil {
			return fmt.Errorf("error updating disbursement status to started for disbursement with id %s: %w", disbursementID, err)
		}

		return nil
	})
}
