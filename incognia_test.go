package incognia

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	clientID       string = "client-id"
	clientSecret   string = "client-secret"
	token          string = "some-token"
	tokenExpiresIn string = "500"
)

var (
	installationId                               = "installation-id"
	sessionToken                                 = "session-token"
	shouldEval               bool                = true
	shouldNotEval            bool                = false
	emptyQueryString         map[string][]string = nil
	queryStringWithFalseEval                     = map[string][]string{"eval": []string{"false"}}
	queryStringWithTrueEval                      = map[string][]string{"eval": []string{"true"}}
	signupAssessmentFixture                      = &SignupAssessment{
		ID:             "some-id",
		DeviceID:       "some-device-id",
		RequestID:      "some-request-id",
		RiskAssessment: LowRisk,
		Reasons:        []Reason{{Code: "mpos_fraud", Source: "global"}, {Code: "mpos_fraud", Source: "local"}},
		Evidence: Evidence{
			"device_model":                 "Moto Z2 Play",
			"geocode_quality":              "good",
			"address_quality":              "good",
			"address_match":                "street",
			"location_events_near_address": 38.0,
			"location_events_quantity":     288.0,
			"location_services": map[string]interface{}{
				"location_permission_enabled": true,
				"location_sensors_enabled":    true,
			},
			"device_integrity": map[string]interface{}{
				"probable_root":       false,
				"emulator":            false,
				"gps_spoofing":        false,
				"from_official_store": true,
			},
		},
	}
	postSignupRequestBodyFixture = &postAssessmentRequestBody{
		InstallationID: installationId,
		AddressLine:    "address line",
		StructuredAddress: &StructuredAddress{
			Locale:       "locale",
			CountryName:  "country-name",
			CountryCode:  "country-code",
			State:        "state",
			City:         "city",
			Borough:      "borough",
			Neighborhood: "neighborhood",
			Street:       "street",
			Number:       "number",
			Complements:  "complements",
			PostalCode:   "postalcode",
		},
		Coordinates: &Coordinates{
			Lat: -23.561414,
			Lng: -46.6558819,
		},
	}
	postSignupRequestBodyRequiredFieldsFixture = &postAssessmentRequestBody{
		InstallationID: installationId,
	}
	addressFixture = &Address{
		Coordinates:       postSignupRequestBodyFixture.Coordinates,
		StructuredAddress: postSignupRequestBodyFixture.StructuredAddress,
		AddressLine:       postSignupRequestBodyFixture.AddressLine,
	}
	postFeedbackRequestBodyFixture = &postFeedbackRequestBody{
		Event:          SignupAccepted,
		Timestamp:      time.Now().UnixNano() / 1000000,
		InstallationID: "some-installation-id",
		LoginID:        "some-login-id",
		PaymentID:      "some-payment-id",
		SignupID:       "some-signup-id",
		AccountID:      "some-account-id",
		ExternalID:     "some-external-id",
	}
	postFeedbackRequestBodyRequiredFieldsFixture = &postFeedbackRequestBody{
		Event:     SignupAccepted,
		Timestamp: time.Now().UnixNano() / 1000000,
	}
	feedbackIdentifiersFixture = &FeedbackIdentifiers{
		InstallationID: "some-installation-id",
		LoginID:        "some-login-id",
		PaymentID:      "some-payment-id",
		SignupID:       "some-signup-id",
		AccountID:      "some-account-id",
		ExternalID:     "some-external-id",
	}
	emptyTransactionAssessmentFixture = &TransactionAssessment{}
	transactionAssessmentFixture      = &TransactionAssessment{
		ID:             "some-id",
		DeviceID:       "some-device-id",
		RiskAssessment: LowRisk,
		Reasons:        []Reason{{Code: "mpos_fraud", Source: "global"}, {Code: "mpos_fraud", Source: "local"}},
		Evidence: Evidence{
			"device_model":                 "Moto Z2 Play",
			"geocode_quality":              "good",
			"address_quality":              "good",
			"address_match":                "street",
			"location_events_near_address": 38.0,
			"location_events_quantity":     288.0,
			"location_services": map[string]interface{}{
				"location_permission_enabled": true,
				"location_sensors_enabled":    true,
			},
			"device_integrity": map[string]interface{}{
				"probable_root":       false,
				"emulator":            false,
				"gps_spoofing":        false,
				"from_official_store": true,
			},
		},
	}
	postPaymentRequestBodyFixture = &postTransactionRequestBody{
		InstallationID: &installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		Type:           paymentType,
		Addresses: []*TransactionAddress{
			{
				Type: Billing,
				StructuredAddress: &StructuredAddress{
					Locale:       "locale",
					CountryName:  "country-name",
					CountryCode:  "country-code",
					State:        "state",
					City:         "city",
					Borough:      "borough",
					Neighborhood: "neighborhood",
					Street:       "street",
					Number:       "number",
					Complements:  "complements",
					PostalCode:   "postalcode",
				},
				AddressLine: "address line",
				Coordinates: &Coordinates{
					Lat: -23.561414,
					Lng: -46.6558819,
				},
			},
		},
		PaymentValue: &PaymentValue{
			Amount:   55.02,
			Currency: "BRL",
		},
		PaymentMethods: []*PaymentMethod{
			{
				Type:       CreditCard,
				Identifier: "credit-card-hash-123",
				CreditCard: &CardInfo{
					Bin:            "29282",
					LastFourDigits: "2222",
					ExpiryYear:     "2020",
					ExpiryMonth:    "10",
				},
			},
		},
	}
	postPaymentRequestBodyRequiredFieldsFixture = &postTransactionRequestBody{
		InstallationID: &installationId,
		AccountID:      "account-id",
		Type:           paymentType,
	}
	paymentFixture = &Payment{
		InstallationID: installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		Addresses: []*TransactionAddress{
			{
				Type: Billing,
				StructuredAddress: &StructuredAddress{
					Locale:       "locale",
					CountryName:  "country-name",
					CountryCode:  "country-code",
					State:        "state",
					City:         "city",
					Borough:      "borough",
					Neighborhood: "neighborhood",
					Street:       "street",
					Number:       "number",
					Complements:  "complements",
					PostalCode:   "postalcode",
				},
				AddressLine: "address line",
				Coordinates: &Coordinates{
					Lat: -23.561414,
					Lng: -46.6558819,
				},
			},
		},
		Value: &PaymentValue{
			Amount:   55.02,
			Currency: "BRL",
		},
		Methods: []*PaymentMethod{
			{
				Type:       CreditCard,
				Identifier: "credit-card-hash-123",
				CreditCard: &CardInfo{
					Bin:            "29282",
					LastFourDigits: "2222",
					ExpiryYear:     "2020",
					ExpiryMonth:    "10",
				},
			},
		},
	}
	paymentFixtureRequiredFields = &Payment{
		InstallationID: installationId,
		AccountID:      "account-id",
	}
	simplePaymentFixtureWithShouldEval = &Payment{
		InstallationID: installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		Eval:           &shouldEval,
	}
	simplePaymentFixtureWithShouldNotEval = &Payment{
		InstallationID: installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		Eval:           &shouldNotEval,
	}
	postSimplePaymentRequestBodyFixture = &postTransactionRequestBody{
		InstallationID: &installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		Type:           paymentType,
	}
	loginFixture = &Login{
		InstallationID:          &installationId,
		AccountID:               "account-id",
		ExternalID:              "external-id",
		PolicyID:                "policy-id",
		PaymentMethodIdentifier: "payment-method-identifier",
	}
	loginFixtureWithShouldEval = &Login{
		InstallationID:          &installationId,
		AccountID:               "account-id",
		ExternalID:              "external-id",
		PolicyID:                "policy-id",
		PaymentMethodIdentifier: "payment-method-identifier",
		Eval:                    &shouldEval,
	}
	loginFixtureWithShouldNotEval = &Login{
		InstallationID: &installationId,
		AccountID:      "account-id",
		ExternalID:     "external-id",
		PolicyID:       "policy-id",
		Eval:           &shouldNotEval,
	}
	loginWebFixture = &Login{
		AccountID:               "account-id",
		ExternalID:              "external-id",
		PolicyID:                "policy-id",
		PaymentMethodIdentifier: "payment-method-identifier",
		SessionToken:            &sessionToken,
	}
	postLoginRequestBodyFixture = &postTransactionRequestBody{
		InstallationID:          &installationId,
		AccountID:               "account-id",
		ExternalID:              "external-id",
		PolicyID:                "policy-id",
		PaymentMethodIdentifier: "payment-method-identifier",
		Type:                    loginType,
	}
	postLoginWebRequestBodyFixture = &postTransactionRequestBody{
		AccountID:               "account-id",
		ExternalID:              "external-id",
		PolicyID:                "policy-id",
		PaymentMethodIdentifier: "payment-method-identifier",
		Type:                    loginType,
		SessionToken:            &sessionToken,
	}
)

type PanickingTokenProvider struct {
	panicString string
}

func (tokenProvider PanickingTokenProvider) GetToken() (Token, error) {
	panic(tokenProvider.panicString)
}

type IncogniaTestSuite struct {
	suite.Suite

	client      *Client
	token       string
	tokenServer *httptest.Server
}

func (suite *IncogniaTestSuite) SetupTest() {
	client, _ := New(&IncogniaClientConfig{ClientID: clientID, ClientSecret: clientSecret})
	suite.client = client

	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	suite.client.tokenProvider.(*AutoRefreshTokenProvider).tokenClient.tokenEndpoint = tokenServer.URL

	suite.client.endpoints.Token = tokenServer.URL
	suite.token = token
	suite.tokenServer = tokenServer
}

func (suite *IncogniaTestSuite) TearDownTest() {
	defer suite.tokenServer.Close()
}

func (suite *IncogniaTestSuite) TestManualRefreshTokenProviderErrorTokenNotFound() {
	tokenProvider := NewManualRefreshTokenProvider(NewTokenClient(&TokenClientConfig{ClientID: clientID, ClientSecret: clientSecret}))
	client, _ := New(&IncogniaClientConfig{ClientID: clientID, ClientSecret: clientSecret, TokenProvider: tokenProvider})

	_, err := client.GetSignupAssessment("any-signup-id")
	suite.EqualError(err, ErrTokenNotFound.Error())

	_, err = client.RegisterLogin(loginFixture)
	suite.EqualError(err, ErrTokenNotFound.Error())

	_, err = client.RegisterPayment(paymentFixture)
	suite.EqualError(err, ErrTokenNotFound.Error())

	timestamp := time.Unix(0, postFeedbackRequestBodyFixture.Timestamp*int64(1000000))
	err = client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.EqualError(err, ErrTokenNotFound.Error())
}

func (suite *IncogniaTestSuite) TestManualRefreshTokenProviderSuccess() {
	tokenProvider := NewManualRefreshTokenProvider(NewTokenClient(&TokenClientConfig{ClientID: clientID, ClientSecret: clientSecret}))
	tokenServer := mockTokenEndpoint(token, tokenExpiresIn)
	defer tokenServer.Close()

	tokenProvider.tokenClient.tokenEndpoint = tokenServer.URL
	client, _ := New(&IncogniaClientConfig{ClientID: clientID, ClientSecret: clientSecret, TokenProvider: tokenProvider})

	tokenProvider.Refresh()

	suite.client = client
	signupID := "signup-id"

	signupServer := suite.mockGetSignupsEndpoint(token, signupID, signupAssessmentFixture)
	defer signupServer.Close()
	_, err := client.GetSignupAssessment(signupID)
	suite.NoError(err)

	loginServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer loginServer.Close()
	_, err = client.RegisterLogin(loginFixture)
	suite.NoError(err)

	paymentServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer paymentServer.Close()
	_, err = client.RegisterPayment(paymentFixture)
	suite.NoError(err)

	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()
	timestamp := time.Unix(0, postFeedbackRequestBodyFixture.Timestamp*int64(1000000))
	err = client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestSuccessGetSignupAssessment() {
	signupID := "signup-id"
	signupServer := suite.mockGetSignupsEndpoint(token, signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupID)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessGetSignupAssessmentAfterTokenExpiration() {
	signupID := "signup-id"
	signupServer := suite.mockGetSignupsEndpoint(token, signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupID)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	response, err = suite.client.GetSignupAssessment(signupID)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}
func (suite *IncogniaTestSuite) TestGetSignupAssessmentEmptySignupId() {
	response, err := suite.client.GetSignupAssessment("")
	suite.EqualError(err, ErrMissingSignupID.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenGetSignupAssessment() {
	signupID := "signup-id"
	signupServer := suite.mockGetSignupsEndpoint("some-other-token", signupID, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.GetSignupAssessment(signupID)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestGetSignupAssessmentErrors() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		suite.client.endpoints.Signups = statusServer.URL

		response, err := suite.client.GetSignupAssessment("any-signup-id")
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterSignup() {
	signupServer := suite.mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterSignupNilOptional() {
	signupServer := suite.mockPostSignupsEndpoint(token, postSignupRequestBodyRequiredFieldsFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyRequiredFieldsFixture.InstallationID, nil)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterSignupAfterTokenExpiration() {
	signupServer := suite.mockPostSignupsEndpoint(token, postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	response, err = suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.NoError(err)
	suite.Equal(signupAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestRegisterSignupEmptyInstallationId() {
	response, err := suite.client.RegisterSignup("", &Address{})
	suite.EqualError(err, ErrMissingInstallationID.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterSignup() {
	signupServer := suite.mockPostSignupsEndpoint("some-other-token", postSignupRequestBodyFixture, signupAssessmentFixture)
	defer signupServer.Close()

	response, err := suite.client.RegisterSignup(postSignupRequestBodyFixture.InstallationID, addressFixture)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestRegisterSignupErrors() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		suite.client.endpoints.Signups = statusServer.URL

		response, err := suite.client.RegisterSignup("any-signup-id", &Address{})
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterFeedback() {
	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()

	timestamp := time.Unix(0, postFeedbackRequestBodyFixture.Timestamp*int64(1000000))
	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterFeedbackNilOptional() {
	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestBodyRequiredFieldsFixture)
	defer feedbackServer.Close()

	timestamp := time.Unix(0, postFeedbackRequestBodyRequiredFieldsFixture.Timestamp*int64(1000000))
	err := suite.client.RegisterFeedback(postFeedbackRequestBodyRequiredFieldsFixture.Event, &timestamp, nil)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterFeedbackAfterTokenExpiration() {
	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()

	timestamp := time.Unix(0, postFeedbackRequestBodyFixture.Timestamp*int64(1000000))
	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.NoError(err)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	err = suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.NoError(err)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterFeedback() {
	feedbackServer := suite.mockFeedbackEndpoint("some-other-token", postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()

	timestamp := time.Unix(0, postFeedbackRequestBodyFixture.Timestamp*int64(1000000))
	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestErrorRegisterFeedbackInvalidFeedbackType() {
	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()

	timestamp := time.Unix(0, postFeedbackRequestBodyFixture.Timestamp*int64(1000000))
	err := suite.client.RegisterFeedback("invalid-type", &timestamp, feedbackIdentifiersFixture)
	suite.EqualError(err, ErrInvalidFeedbackType.Error())
}

func (suite *IncogniaTestSuite) TestErrorRegisterFeedbackNilTimestamp() {
	feedbackServer := suite.mockFeedbackEndpoint(token, postFeedbackRequestBodyFixture)
	defer feedbackServer.Close()

	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, nil, feedbackIdentifiersFixture)
	suite.EqualError(err, ErrMissingTimestamp.Error())
}

func (suite *IncogniaTestSuite) TestErrorsRegisterFeedback() {
	timestamp := time.Unix(0, postFeedbackRequestBodyFixture.Timestamp*int64(1000000))

	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		suite.client.endpoints.Feedback = statusServer.URL

		err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPayment() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixture)

	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentNilOptional() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyRequiredFieldsFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixtureRequiredFields)

	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentAfterTokenExpiration() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postPaymentRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	response, err = suite.client.RegisterPayment(paymentFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestRegisterPaymentNilPayment() {
	response, err := suite.client.RegisterPayment(nil)
	suite.EqualError(err, ErrMissingPayment.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestRegisterPaymentEmptyInstallationId() {
	response, err := suite.client.RegisterPayment(&Payment{AccountID: "some-account-id"})
	suite.EqualError(err, ErrMissingInstallationID.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestRegisterPaymentEmptyAccountId() {
	response, err := suite.client.RegisterPayment(&Payment{InstallationID: "some-installation-id"})
	suite.EqualError(err, ErrMissingAccountID.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterPayment() {
	transactionServer := suite.mockPostTransactionsEndpoint("some-other-token", postPaymentRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(paymentFixture)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestRegisterPaymentErrors() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		suite.client.endpoints.Transactions = statusServer.URL

		response, err := suite.client.RegisterPayment(paymentFixture)
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentWithEval() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postSimplePaymentRequestBodyFixture, transactionAssessmentFixture, queryStringWithTrueEval)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(simplePaymentFixtureWithShouldEval)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterPaymentWithFalseEval() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postSimplePaymentRequestBodyFixture, transactionAssessmentFixture, queryStringWithFalseEval)
	defer transactionServer.Close()

	response, err := suite.client.RegisterPayment(simplePaymentFixtureWithShouldNotEval)
	suite.NoError(err)
	suite.Equal(emptyTransactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLogin() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginWithEval() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, queryStringWithTrueEval)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixtureWithShouldEval)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginWithFalseEval() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, queryStringWithFalseEval)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixtureWithShouldNotEval)
	suite.NoError(err)
	suite.Equal(emptyTransactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginWeb() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginWebRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.registerLogin(loginWebFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestSuccessRegisterLoginAfterTokenExpiration() {
	transactionServer := suite.mockPostTransactionsEndpoint(token, postLoginRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)

	token, _ := suite.client.tokenProvider.GetToken()
	token.(*accessToken).ExpiresIn = 0

	response, err = suite.client.RegisterLogin(loginFixture)
	suite.NoError(err)
	suite.Equal(transactionAssessmentFixture, response)
}

func (suite *IncogniaTestSuite) TestRegisterLoginNilLogin() {
	response, err := suite.client.RegisterLogin(nil)
	suite.EqualError(err, ErrMissingLogin.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestRegisterLoginNullInstallationIdAndSessionToken() {
	response, err := suite.client.RegisterLogin(&Login{AccountID: "some-account-id"})
	suite.EqualError(err, ErrMissingInstallationIDOrSessionToken.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestRegisterLoginEmptyAccountId() {
	response, err := suite.client.RegisterLogin(&Login{InstallationID: &installationId})
	suite.EqualError(err, ErrMissingAccountID.Error())
	suite.Nil(response)
}

func (suite *IncogniaTestSuite) TestForbiddenRegisterLogin() {
	transactionServer := suite.mockPostTransactionsEndpoint("some-other-token", postLoginRequestBodyFixture, transactionAssessmentFixture, emptyQueryString)
	defer transactionServer.Close()

	response, err := suite.client.RegisterLogin(loginFixture)
	suite.Nil(response)
	suite.EqualError(err, "403 Forbidden")
}

func (suite *IncogniaTestSuite) TestUnauthorizedTokenGeneration() {
	tokenServer := suite.mockTokenEndpointUnauthorized()
	suite.client.tokenProvider.(*AutoRefreshTokenProvider).tokenClient.tokenEndpoint = tokenServer.URL
	defer tokenServer.Close()

	responsePayment, err := suite.client.RegisterPayment(paymentFixture)
	suite.Nil(responsePayment)
	suite.EqualError(err, ErrInvalidCredentials.Error())

	responseLogin, err := suite.client.RegisterLogin(loginFixture)
	suite.Nil(responseLogin)
	suite.EqualError(err, ErrInvalidCredentials.Error())

	responseSignUp, err := suite.client.RegisterSignup(installationId, addressFixture)
	suite.Nil(responseSignUp)
	suite.EqualError(err, ErrInvalidCredentials.Error())

	timestamp := time.Unix(0, postFeedbackRequestBodyFixture.Timestamp*int64(1000000))
	err = suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.EqualError(err, ErrInvalidCredentials.Error())
}

func (suite *IncogniaTestSuite) TestRegisterLoginErrors() {
	errors := []int{http.StatusBadRequest, http.StatusInternalServerError}
	for _, status := range errors {
		statusServer := mockStatusServer(status)
		suite.client.endpoints.Transactions = statusServer.URL

		response, err := suite.client.RegisterLogin(loginFixture)
		suite.Nil(response)
		suite.Contains(err.Error(), strconv.Itoa(status))
	}
}

func (suite *IncogniaTestSuite) TestPanic() {
	defer func() { suite.Nil(recover()) }()

	panicString := "error getting token"
	suite.client.tokenProvider = &PanickingTokenProvider{panicString: panicString}

	suite.client.RegisterLogin(loginFixture)
	timestamp := time.Unix(0, postFeedbackRequestBodyFixture.Timestamp*int64(1000000))
	err := suite.client.RegisterFeedback(postFeedbackRequestBodyFixture.Event, &timestamp, feedbackIdentifiersFixture)
	suite.Equal(err.Error(), panicString)
	_, err = suite.client.RegisterSignup("some-installationId", addressFixture)
	suite.Equal(err.Error(), panicString)
	_, err = suite.client.GetSignupAssessment("some-signup-id")
	suite.Equal(err.Error(), panicString)
	_, err = suite.client.RegisterPayment(paymentFixture)
	suite.Equal(err.Error(), panicString)
}

func TestIncogniaTestSuite(t *testing.T) {
	suite.Run(t, new(IncogniaTestSuite))
}

func (suite *IncogniaTestSuite) mockFeedbackEndpoint(expectedToken string, expectedBody *postFeedbackRequestBody) *httptest.Server {
	feedbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		if !isRequestAuthorized(r, expectedToken) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var requestBody postFeedbackRequestBody
		json.NewDecoder(r.Body).Decode(&requestBody)

		if reflect.DeepEqual(&requestBody, expectedBody) {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))

	suite.client.endpoints.Feedback = feedbackServer.URL

	return feedbackServer
}

func mockStatusServer(statusCode int) *httptest.Server {
	statusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(statusCode)
	}))

	return statusServer
}

func (suite *IncogniaTestSuite) mockTokenEndpointUnauthorized() *httptest.Server {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
	}))

	return tokenServer
}

func (suite *IncogniaTestSuite) mockPostTransactionsEndpoint(expectedToken string, expectedBody *postTransactionRequestBody, expectedResponse *TransactionAssessment, expectedQueryString map[string][]string) *httptest.Server {
	transactionsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		if !isRequestAuthorized(r, expectedToken) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		requestQueryString := r.URL.Query()
		for parameter := range requestQueryString {
			suite.Equal(expectedQueryString[parameter], requestQueryString[parameter])
		}

		requestEvalQueryString := requestQueryString["eval"]
		if requestEvalQueryString != nil && requestEvalQueryString[0] == "false" {
			res, _ := json.Marshal(emptyTransactionAssessmentFixture)
			w.Write(res)
			return
		}

		var requestBody postTransactionRequestBody
		json.NewDecoder(r.Body).Decode(&requestBody)

		if reflect.DeepEqual(&requestBody, expectedBody) {
			res, _ := json.Marshal(expectedResponse)
			w.Write(res)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	}))

	suite.client.endpoints.Transactions = transactionsServer.URL

	return transactionsServer
}

func (suite *IncogniaTestSuite) mockPostSignupsEndpoint(expectedToken string, expectedBody *postAssessmentRequestBody, expectedResponse *SignupAssessment) *httptest.Server {
	signupsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		if !isRequestAuthorized(r, expectedToken) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var requestBody postAssessmentRequestBody
		json.NewDecoder(r.Body).Decode(&requestBody)

		if reflect.DeepEqual(&requestBody, expectedBody) {
			res, _ := json.Marshal(expectedResponse)
			w.Write(res)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	}))

	suite.client.endpoints.Signups = signupsServer.URL

	return signupsServer
}

func (suite *IncogniaTestSuite) mockGetSignupsEndpoint(expectedToken, expectedSignupID string, expectedResponse *SignupAssessment) *httptest.Server {
	getSignupsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		if !isRequestAuthorized(r, expectedToken) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		defer r.Body.Close()

		splitUrl := strings.Split(r.URL.RequestURI(), "/")
		requestSignupID := splitUrl[len(splitUrl)-1]

		if requestSignupID == expectedSignupID {
			res, _ := json.Marshal(expectedResponse)
			w.Write(res)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))

	suite.client.endpoints.Signups = getSignupsServer.URL

	return getSignupsServer
}

func isRequestAuthorized(request *http.Request, expectedToken string) bool {
	tokenType, token := readAuthorizationHeader(request)

	return token == expectedToken && tokenType == "Bearer"
}

func readAuthorizationHeader(request *http.Request) (string, string) {
	receivedToken := strings.Split(request.Header.Get("Authorization"), " ")
	tokenType := receivedToken[0]
	token := receivedToken[1]

	return tokenType, token
}

func mockTokenEndpoint(expectedToken string, expiresIn string) *httptest.Server {
	tokenResponse := map[string]string{
		"access_token": expectedToken,
		"expires_in":   expiresIn,
		"token_type":   "Bearer",
	}

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		username, password, ok := r.BasicAuth()

		if !ok || username != clientID || password != clientSecret {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		res, _ := json.Marshal(tokenResponse)
		w.Write(res)
	}))

	return tokenServer
}
