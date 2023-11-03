package tenant

import (
	"testing"

	"github.com/stellar/go/keypair"
	"github.com/stretchr/testify/assert"
)

func Test_TenantUpdate_Validate(t *testing.T) {
	t.Run("invalid values", func(t *testing.T) {
		tu := TenantUpdate{}
		err := tu.Validate()
		assert.EqualError(t, err, "tenant ID is required")

		tu.ID = "abc"
		err = tu.Validate()
		assert.EqualError(t, err, "provide at least one field to be updated")

		esType := EmailSenderType("invalid")
		tu.EmailSenderType = &esType
		err = tu.Validate()
		assert.EqualError(t, err, `invalid email sender type: invalid email sender type "invalid"`)

		tu.EmailSenderType = nil
		smsSenderType := SMSSenderType("invalid")
		tu.SMSSenderType = &smsSenderType
		err = tu.Validate()
		assert.EqualError(t, err, `invalid SMS sender type: invalid sms sender type "invalid"`)

		tu.SMSSenderType = nil
		addr := "invalid"
		tu.SEP10SigningPublicKey = &addr
		err = tu.Validate()
		assert.EqualError(t, err, "invalid SEP-10 signing public key")

		tu.SEP10SigningPublicKey = nil
		tu.DistributionPublicKey = &addr
		err = tu.Validate()
		assert.EqualError(t, err, "invalid distribution public key")

		tu.DistributionPublicKey = nil
		u := "inv@lid$"
		tu.BaseURL = &u
		err = tu.Validate()
		assert.EqualError(t, err, "invalid base URL")

		tu.BaseURL = nil
		tu.SDPUIBaseURL = &u
		err = tu.Validate()
		assert.EqualError(t, err, "invalid SDP UI base URL")

		tu.SDPUIBaseURL = nil
		tu.CORSAllowedOrigins = []string{"inv@lid$"}
		err = tu.Validate()
		assert.EqualError(t, err, `invalid CORS allowed origin url: "inv@lid$"`)

		tu.CORSAllowedOrigins = nil
		tenantStatus := TenantStatus("invalid")
		tu.Status = &tenantStatus
		err = tu.Validate()
		assert.EqualError(t, err, `invalid tenant status: "invalid"`)
	})

	t.Run("valid values", func(t *testing.T) {
		tu := TenantUpdate{
			ID:                    "abc",
			EmailSenderType:       &AWSEmailSenderType,
			SMSSenderType:         &TwilioSMSSenderType,
			SEP10SigningPublicKey: &[]string{keypair.MustRandom().Address()}[0],
			DistributionPublicKey: &[]string{keypair.MustRandom().Address()}[0],
			EnableMFA:             &[]bool{true}[0],
			EnableReCAPTCHA:       &[]bool{true}[0],
			CORSAllowedOrigins:    []string{"https://myorg.sdp.io", "https://myorg-dev.sdp.io"},
			BaseURL:               &[]string{"https://myorg.backend.io"}[0],
			SDPUIBaseURL:          &[]string{"https://myorg.frontend.io"}[0],
			Status:                &[]TenantStatus{ProvisionedTenantStatus}[0],
		}
		err := tu.Validate()
		assert.NoError(t, err)
	})
}

func Test_TenantUpdate_areAllFieldsEmpty(t *testing.T) {
	tu := TenantUpdate{}
	assert.True(t, tu.areAllFieldsEmpty())
	tu.SDPUIBaseURL = &[]string{"https://myorg.backend.io"}[0]
	assert.False(t, tu.areAllFieldsEmpty())
}

func Test_ParseEmailSenderType(t *testing.T) {
	est, err := ParseEmailSenderType("invalid")
	assert.EqualError(t, err, `invalid email sender type "invalid"`)
	assert.Empty(t, est)

	est, err = ParseEmailSenderType("aws_email")
	assert.EqualError(t, err, `invalid email sender type "aws_email"`)
	assert.Empty(t, est)

	est, err = ParseEmailSenderType("AWS_EMAIL")
	assert.NoError(t, err)
	assert.Equal(t, AWSEmailSenderType, est)
}

func Test_ParseSMSSenderType(t *testing.T) {
	sst, err := ParseSMSSenderType("invalid")
	assert.EqualError(t, err, `invalid sms sender type "invalid"`)
	assert.Empty(t, sst)

	sst, err = ParseSMSSenderType("twilio_sms")
	assert.EqualError(t, err, `invalid sms sender type "twilio_sms"`)
	assert.Empty(t, sst)

	sst, err = ParseSMSSenderType("TWILIO_SMS")
	assert.NoError(t, err)
	assert.Equal(t, TwilioSMSSenderType, sst)
}

func Test_TenantStatus_IsValid(t *testing.T) {
	testCases := []struct {
		status TenantStatus
		expect bool
	}{
		{
			status: CreatedTenantStatus,
			expect: true,
		},
		{
			status: ProvisionedTenantStatus,
			expect: true,
		},
		{
			status: ActivatedTenantStatus,
			expect: true,
		},
		{
			status: DeactivatedTenantStatus,
			expect: true,
		},
		{
			status: TenantStatus("invalid"),
			expect: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expect, tc.status.IsValid())
	}
}
