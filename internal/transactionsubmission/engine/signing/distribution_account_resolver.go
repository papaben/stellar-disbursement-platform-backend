package signing

import (
	"context"
	"fmt"

	"github.com/stellar/go/strkey"
	"github.com/stellar/stellar-disbursement-platform-backend/db"
	"github.com/stellar/stellar-disbursement-platform-backend/stellar-multitenant/pkg/tenant"
)

var ErrDistributionAccountIsEmpty = fmt.Errorf("distribution account is empty")

// DistributionAccountResolver is an interface that provides the distribution iven the provided keyword.
//
//go:generate mockery --name=DistributionAccountResolver --case=underscore --structname=MockDistributionAccountResolver
type DistributionAccountResolver interface {
	DistributionAccount(ctx context.Context, tenantID string) (string, error)
	DistributionAccountFromContext(ctx context.Context) (string, error)
	HostDistributionAccount() string
}

type DistributionAccountResolverOptions struct {
	AdminDBConnectionPool            db.DBConnectionPool
	HostDistributionAccountPublicKey string
}

func (c DistributionAccountResolverOptions) Validate() error {
	if c.AdminDBConnectionPool == nil {
		return fmt.Errorf("AdminDBConnectionPool cannot be nil")
	}

	if c.HostDistributionAccountPublicKey == "" {
		return fmt.Errorf("HostDistributionAccountPublicKey cannot be empty")
	}
	if !strkey.IsValidEd25519PublicKey(c.HostDistributionAccountPublicKey) {
		return fmt.Errorf("HostDistributionAccountPublicKey is not a valid ed25519 public key")
	}

	return nil
}

func NewDistributionAccountResolver(config DistributionAccountResolverOptions) (DistributionAccountResolver, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validating config in NewDistributionAccountResolver: %w", err)
	}

	return &DistributionAccountResolverImpl{
		tenantManager:                 tenant.NewManager(tenant.WithDatabase(config.AdminDBConnectionPool)),
		hostDistributionAccountPubKey: config.HostDistributionAccountPublicKey,
	}, nil
}

var _ DistributionAccountResolver = (*DistributionAccountResolverImpl)(nil)

// DistributionAccountResolverImpl is a DistributionAccountResolver that resolves the distribution account from the database.
type DistributionAccountResolverImpl struct {
	tenantManager                 tenant.ManagerInterface
	hostDistributionAccountPubKey string
}

// DistributionAccount returns the tenant's distribution account stored in the database.
func (r *DistributionAccountResolverImpl) DistributionAccount(ctx context.Context, tenantID string) (string, error) {
	tnt, err := r.tenantManager.GetTenantByID(ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("getting tenant by ID: %w", err)
	}

	if tnt.DistributionAccount == nil {
		return "", ErrDistributionAccountIsEmpty
	}

	return *tnt.DistributionAccount, nil
}

// DistributionAccountFromContext returns the tenant's distribution account from the tenant object stored in the context
// provided.
func (r *DistributionAccountResolverImpl) DistributionAccountFromContext(ctx context.Context) (string, error) {
	tnt, err := tenant.GetTenantFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("getting tenant from context: %w", err)
	}

	if tnt.DistributionAccount == nil {
		return "", ErrDistributionAccountIsEmpty
	}

	return *tnt.DistributionAccount, nil
}

// HostDistributionAccount returns the host distribution account from the database.
func (r *DistributionAccountResolverImpl) HostDistributionAccount() string {
	return r.hostDistributionAccountPubKey
}