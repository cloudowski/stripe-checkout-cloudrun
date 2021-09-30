package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/webhook"
)

func main() {
	checkEnv()

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	http.HandleFunc("/config", handleConfig)
	http.HandleFunc("/checkout-session", handleCheckoutSession)
	http.HandleFunc("/create-checkout-session", handleCreateCheckoutSession)
	// not required in prod
	http.HandleFunc("/webhook", handleWebhook)

	log.Println("server running at 0.0.0.0:8080")
	http.ListenAndServe("0.0.0.0:8080", nil)
}

// ErrorResponseMessage represents the structure of the error
// object sent in failed responses.
type ErrorResponseMessage struct {
	Message string `json:"message"`
}

// ErrorResponse represents the structure of the error object sent
// in failed responses.
type ErrorResponse struct {
	Error *ErrorResponseMessage `json:"error"`
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	// Fetch a price, use it's unit amount and currency
	p, _ := price.Get(
		os.Getenv("PRICE"),
		nil,
	)
	writeJSON(w, struct {
		PublicKey  string `json:"publicKey"`
		UnitAmount int64  `json:"unitAmount"`
		Currency   string `json:"currency"`
	}{
		PublicKey:  os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		UnitAmount: p.UnitAmount,
		Currency:   string(p.Currency),
	})
}

func handleCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	sessionID := r.URL.Query().Get("sessionId")
	s, _ := session.Get(sessionID, nil)
	writeJSON(w, s)
}

func handleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	locale := getEnv("STRIPE_LOCALE", "auto")
	allow_promocodes, _ := strconv.ParseBool(getEnv("STRIPE_ALLOW_PROMOCODES", "false"))
	collect_address := getEnv("STRIPE_COLLECT_ADDRESS", "auto")
	collect_taxid, _ := strconv.ParseBool(getEnv("STRIPE_COLLECT_TAXID", "false"))
	quantity := int64(1)
	url_cancel := getEnv("STRIPE_URL_CANCEL", "https://cloudowski.com")
	url_success := getEnv("STRIPE_URL_SUCCESS", "https://cloudowski.com")

	paymentMethodTypes := strings.Split(os.Getenv("PAYMENT_METHOD_TYPES"), ",")

	// For full details see https://stripe.com/docs/api/checkout/sessions/create

	params := &stripe.CheckoutSessionParams{
		SuccessURL:         stripe.String(url_success),
		CancelURL:          stripe.String(url_cancel),
		PaymentMethodTypes: stripe.StringSlice(paymentMethodTypes),
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(quantity),
				Price:    stripe.String(os.Getenv("PRICE")),
			},
		},
		BillingAddressCollection: stripe.String(collect_address),
		AllowPromotionCodes:      stripe.Bool(allow_promocodes),
		TaxIDCollection:          &stripe.CheckoutSessionTaxIDCollectionParams{Enabled: stripe.Bool(collect_taxid)},
		Locale:                   stripe.String(locale),
	}
	s, err := session.New(params)
	if err != nil {
		http.Error(w, fmt.Sprintf("error while creating session %v", err.Error()), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("ioutil.ReadAll: %v", err)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return
	}

	if event.Type == "checkout.session.completed" {
		fmt.Println("Checkout Session completed!")
	}

	writeJSON(w, nil)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}

func writeJSONError(w http.ResponseWriter, v interface{}, code int) {
	w.WriteHeader(code)
	writeJSON(w, v)
	return
}

func writeJSONErrorMessage(w http.ResponseWriter, message string, code int) {
	resp := &ErrorResponse{
		Error: &ErrorResponseMessage{
			Message: message,
		},
	}
	writeJSONError(w, resp, code)
}

func checkEnv() {
	price := os.Getenv("PRICE")
	if price == "price_12345" || price == "" {
		log.Fatal("You must set a Price ID from your Stripe account. See the README for instructions.")
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
