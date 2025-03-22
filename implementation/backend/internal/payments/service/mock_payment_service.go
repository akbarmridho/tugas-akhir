package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/pkg/errors"
	"golang.org/x/net/http2"
	"net/http"
	"net/url"
	"os"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/payments/entity"
	"tugas-akhir/backend/pkg/mock_payment"
)

type MockPaymentService struct {
	mockPayment *mock_payment.APIClient
	protocol    string
	baseHost    string
}

func NewMockPaymentService(config *config.Config) (*MockPaymentService, error) {
	caCertPool := x509.NewCertPool()

	certData, err := os.ReadFile(config.PaymentCertPath)

	if err != nil {
		return nil, err
	}

	if ok := caCertPool.AppendCertsFromPEM(certData); !ok {
		return nil, errors.WithStack(errors.WithMessage(entity.PaymentServiceInternalError, "failed to append cert to pool"))
	}

	// Create a custom transport with HTTP/2 support
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	// Explicitly enable HTTP/2
	err = http2.ConfigureTransport(transport)
	if err != nil {
		return nil, err
	}

	// Create custom HTTP client with the transport
	httpClient := &http.Client{
		Transport: transport,
	}

	// Configure your API client to use this custom HTTP client
	cfg := mock_payment.NewConfiguration()
	cfg.HTTPClient = httpClient

	mockPayment := mock_payment.NewAPIClient(cfg)

	// Validate base URL
	parsedURL, err := url.Parse(config.PaymentServiceUrl)
	if err != nil {
		return nil, errors.WithStack(errors.WithMessage(entity.PaymentServiceInternalError, "failed to parse payment service url"))
	}

	return &MockPaymentService{
		mockPayment: mockPayment,
		protocol:    parsedURL.Scheme,
		baseHost:    parsedURL.Host + parsedURL.Path,
	}, nil
}

func (s *MockPaymentService) GenerateInvoice(ctx context.Context, payload mock_payment.CreateInvoiceRequest) (*mock_payment.Invoice, error) {
	ctx = context.WithValue(context.Background(), mock_payment.ContextServerVariables, map[string]string{
		"protocol": s.protocol,
		"server":   s.baseHost,
	})

	request := s.mockPayment.DefaultAPI.InvoicesPost(ctx)

	invoiceRequest := mock_payment.NewCreateInvoiceRequest(payload.Amount, payload.ExternalId)
	invoiceRequest.Description = payload.Description

	request = request.CreateInvoiceRequest(*invoiceRequest)

	invoice, _, err := request.Execute()

	if err != nil {
		return nil, err
	}

	return invoice, nil
}
