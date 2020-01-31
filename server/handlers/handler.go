package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mooncascade/toy-banklink/services"
)

type payResponse struct {
	URL string `json:"url"`
}

type preparePaymentResponse struct {
	UUID string `json:"uuid"`
}

type handler struct {
	truelayerService services.Service
}

// Handler interface defining handler methods
type Handler interface {
	PayEndpoint(w http.ResponseWriter, r *http.Request)
	BankCallback(w http.ResponseWriter, r *http.Request)
	GetBanks(w http.ResponseWriter, r *http.Request)
	PreparePayment(w http.ResponseWriter, r *http.Request)
	GetPaymentData(w http.ResponseWriter, r *http.Request)
}

//GetHandler returns handler object
func GetHandler(service services.Service) Handler {
	return handler{
		truelayerService: service,
	}
}

// PayEndpoint parses incoming data and responds with truelayer URL
func (h handler) PayEndpoint(w http.ResponseWriter, r *http.Request) {
	log.Println("Pay endpoint")

	paymentRequest := services.CreatePaymentRequest{
		RedirectURI: "http://localhost:3000/api/callback",
	}

	err := json.NewDecoder(r.Body).Decode(&paymentRequest)
	if err != nil {
		HTTPError("Unable to parse request body: "+err.Error(), http.StatusBadRequest, w)
		return
	}

	// Get Truelayer payment URL
	uri, err := h.truelayerService.RequestPaymentURL(paymentRequest)
	if err != nil {
		HTTPError("Unable to process payment request: "+err.Error(), http.StatusBadRequest, w)
		return
	}

	err = json.NewEncoder(w).Encode(payResponse{
		URL: uri,
	})
	if err != nil {
		HTTPError("Unable to parse response body: "+err.Error(), http.StatusInternalServerError, w)
		return
	}
}

// BankCallback handles callback event from truelayer and redirects user to payment page
func (h handler) BankCallback(w http.ResponseWriter, r *http.Request) {
	// Payment was either cancelled or completed and bank redirects us to this endpoint
	paymentID, found := r.URL.Query()["payment_id"]
	if !found {
		HTTPError("Invalid input: payment_id parameter missing", http.StatusBadRequest, w)
		return
	}

	// Receive payment status
	resp, err := h.truelayerService.RequestPaymentData(paymentID[0])
	if err != nil {
		HTTPError("Unable to process bank callback: "+err.Error(), http.StatusInternalServerError, w)
		return
	}

	http.Redirect(w, r, "http://localhost:80/index.html?uuid="+resp.UUID+"&notify=true", http.StatusFound)
}

// GetBanks responds with a list of banks
func (h handler) GetBanks(w http.ResponseWriter, r *http.Request) {
	var responseBody, err = h.truelayerService.RequestBanks()
	if err != nil {
		HTTPError("Unable to process getting banks: "+err.Error(), http.StatusBadRequest, w)
		return
	}

	_, err = w.Write(responseBody)
	if err != nil {
		HTTPError("Unable to write to response body: "+err.Error(), http.StatusInternalServerError, w)
		return
	}
}

// PreparePayment creates a new payment row in the database
// Returns UUID which can be used to proceed to payment page
func (h handler) PreparePayment(w http.ResponseWriter, r *http.Request) {
	requestBody := services.SavePaymentRequest{}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		HTTPError("Unable to parse request body: "+err.Error(), http.StatusBadRequest, w)
		return
	}

	// Prepares a new payment in the local database only.
	paymentUUID, err := h.truelayerService.PreparePayment(requestBody.ReceiverID, requestBody.Amount)
	if err != nil {
		HTTPError("Unable to process payment preparation: "+err.Error(), http.StatusBadRequest, w)
		return
	}

	err = json.NewEncoder(w).Encode(preparePaymentResponse{
		UUID: paymentUUID,
	})
	if err != nil {
		HTTPError("Unable to parse response body: "+err.Error(), http.StatusInternalServerError, w)
		return
	}
}

// GetPaymentData returns payment data by provided UUID
func (h handler) GetPaymentData(w http.ResponseWriter, r *http.Request) {
	uuid, found := mux.Vars(r)["uuid"]
	if !found {
		HTTPError("Invalid input: uuid path missing", http.StatusBadRequest, w)
		return
	}

	// Get payment data from local database only.
	resp, err := h.truelayerService.GetPaymentData(uuid)
	if err != nil {
		HTTPError("Unable to process getting payment data: "+err.Error(), http.StatusBadRequest, w)
		return
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		HTTPError("Unable to parse response body: "+err.Error(), http.StatusInternalServerError, w)
		return
	}
}
