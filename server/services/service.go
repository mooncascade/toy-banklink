package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mooncascade/toy-banklink/dao"
)

// SavePaymentRequest used to save a new payment to the database
type SavePaymentRequest struct {
	ReceiverID string `json:"receiver_id"`
	Amount     int    `json:"amount"`
}

// CreatePaymentRequest used to make a request truelayer API
type CreatePaymentRequest struct {
	UUID                     string `json:"uuid"`
	Amount                   int    `json:"amount"`
	Currency                 string `json:"currency"`
	BeneficiaryName          string `json:"beneficiary_name"`
	BeneficiaryReference     string `json:"beneficiary_reference"`
	BeneficiarySortCode      string `json:"beneficiary_sort_code"`
	BeneficiaryAccountNumber string `json:"beneficiary_account_number"`
	RemitterReference        string `json:"remitter_reference"`
	RedirectURI              string `json:"redirect_uri"`
	RemitterProviderID       string `json:"remitter_provider_id"`
}

type getAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type createPaymentResponse struct {
	Results []createPaymentResult `json:"results"`
}

type createPaymentResult struct {
	SimpID  string `json:"simp_id"`
	AuthURI string `json:"auth_uri"`
}

type getPaymentResponse struct {
	Results []getPaymentResult `json:"results"`
}

type getPaymentResult struct {
	Status string `json:"status"`
}

type service struct {
	paymentDAO          dao.PaymentDAO
	clientID            string
	clientSecret        string
	authToken           string
	authTokenExpireTime time.Time
	httpClient          http.Client
}

// Service interface that defines the service functions
type Service interface {
	RequestPaymentURL(request CreatePaymentRequest) (string, error)
	RequestPaymentData(paymentID string) (dao.GetPaymentDataResponse, error)
	RequestBanks() ([]byte, error)
	PreparePayment(receiverID string, amount int) (string, error)
	GetPaymentData(uuid string) (dao.GetPaymentDataResponse, error)
}

const (
	trueLayerPayURL   = "https://pay-api.truelayer-sandbox.com"
	trueLayerAuthURL  = "https://auth.truelayer-sandbox.com"
	httpClientTimeout = 20
)

// GetService returns service object
func GetService(dao dao.PaymentDAO, clientID, clientSecret string) Service {
	return &service{
		paymentDAO:   dao,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: http.Client{
			Timeout: httpClientTimeout * time.Second,
		},
	}
}

// RequestPaymentUrl returns Truelayer payment url that can be used for payment
func (s *service) RequestPaymentURL(request CreatePaymentRequest) (string, error) {
	token, err := s.GetAccessToken()
	if err != nil {
		return "", err
	}

	uri, err := s.createPayment(token, request.UUID, request)
	if err != nil {
		return "", err
	}

	return uri, nil
}

// RequestPaymentData requests for payment data from Truelayer API
func (s *service) RequestPaymentData(paymentID string) (dao.GetPaymentDataResponse, error) {
	var response getPaymentResponse

	req, err := http.NewRequest("GET", trueLayerPayURL+"/single-immediate-payments/"+paymentID, nil)
	if err != nil {
		return dao.GetPaymentDataResponse{}, err
	}

	token, err := s.GetAccessToken()
	if err != nil {
		return dao.GetPaymentDataResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return dao.GetPaymentDataResponse{}, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return dao.GetPaymentDataResponse{}, err
	}

	paymentDetails, err := s.paymentDAO.GetPaymentByTruelayerID(paymentID)
	if err != nil {
		return dao.GetPaymentDataResponse{}, err
	}

	err = s.paymentDAO.UpdatePaymentStatus(paymentID, response.Results[0].Status)
	if err != nil {
		return dao.GetPaymentDataResponse{}, err
	}

	return paymentDetails, nil
}

// RequestBanks requests for a list of existing banks from Truelayer API
func (s service) RequestBanks() ([]byte, error) {
	resp, err := http.Get(trueLayerPayURL + "/providers?capability=SingleImmediatePayment")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

// PreparePayment prepares a new payment to be used for payment process by Truelayer
func (s service) PreparePayment(receiverID string, amount int) (string, error) {
	uuid, err := s.paymentDAO.InsertPayment(receiverID, amount)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

// GetPaymentData returns existing payment data from the database
func (s service) GetPaymentData(uuid string) (dao.GetPaymentDataResponse, error) {
	resp, err := s.paymentDAO.GetPayment(uuid)
	if err != nil {
		return dao.GetPaymentDataResponse{}, err
	}

	return resp, nil
}

func (s *service) SetAccessToken() (string, error) {
	var (
		data     = url.Values{}
		response getAccessTokenResponse
	)

	data.Set("scope", "payments")
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)
	data.Set("grant_type", "client_credentials")

	resp, err := http.Post(
		trueLayerAuthURL+"/connect/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	s.authTokenExpireTime = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
	s.authToken = response.AccessToken

	return response.AccessToken, nil
}

// GetAccessToken checks if access token is not expired and returns it.
// If access token is expired then asks for a new one and returns that one instead.
func (s *service) GetAccessToken() (string, error) {
	if diff := time.Until(s.authTokenExpireTime); diff <= 0 {
		newToken, err := s.SetAccessToken()
		if err != nil {
			return "", nil
		}

		return newToken, nil
	}

	return s.authToken, nil
}

func (s service) createPayment(token, uuid string, data CreatePaymentRequest) (string, error) {
	var (
		response createPaymentResponse
	)

	jsondata, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", trueLayerPayURL+"/single-immediate-payments", bytes.NewBuffer(jsondata))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var text, readErr = ioutil.ReadAll(resp.Body)
		if readErr != nil {
			return "", readErr
		}

		log.Println(string(text))

		return "", errors.New("truelayer API did not respond with code 200")
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	err = s.paymentDAO.MapToTruelayer(uuid, response.Results[0].SimpID)
	if err != nil {
		return "", err
	}

	return response.Results[0].AuthURI, nil
}
