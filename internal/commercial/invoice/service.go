// Package invoice provides invoice management functionality.
package invoice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"v/internal/database/repository"
	"v/internal/logger"
)

// Common errors
var (
	ErrInvoiceNotFound = errors.New("invoice not found")
	ErrOrderNotFound   = errors.New("order not found")
	ErrInvoiceExists   = errors.New("invoice already exists for this order")
	ErrInvalidInvoice  = errors.New("invalid invoice data")
)

// Invoice represents an invoice.
type Invoice struct {
	ID        int64       `json:"id"`
	InvoiceNo string      `json:"invoice_no"`
	OrderID   int64       `json:"order_id"`
	UserID    int64       `json:"user_id"`
	Amount    int64       `json:"amount"`
	Content   *Content    `json:"content"`
	PDFPath   string      `json:"pdf_path"`
	CreatedAt string      `json:"created_at"`
}

// Content represents invoice content with line items.
type Content struct {
	CompanyName    string     `json:"company_name"`
	CompanyAddress string     `json:"company_address"`
	TaxID          string     `json:"tax_id"`
	CustomerName   string     `json:"customer_name"`
	CustomerEmail  string     `json:"customer_email"`
	Items          []LineItem `json:"items"`
	Subtotal       int64      `json:"subtotal"`
	Discount       int64      `json:"discount"`
	Total          int64      `json:"total"`
}

// LineItem represents a line item in an invoice.
type LineItem struct {
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	Amount      int64  `json:"amount"`
}

// Config holds invoice service configuration.
type Config struct {
	CompanyName    string `json:"company_name"`
	CompanyAddress string `json:"company_address"`
	TaxID          string `json:"tax_id"`
	NumberFormat   string `json:"number_format"` // e.g., "INV-{YYYY}{MM}-{SEQ}"
	PDFStoragePath string `json:"pdf_storage_path"`
}

// DefaultConfig returns default configuration.
func DefaultConfig() *Config {
	return &Config{
		CompanyName:    "V Panel",
		CompanyAddress: "",
		TaxID:          "",
		NumberFormat:   "INV-%s-%04d",
		PDFStoragePath: "data/invoices",
	}
}


// Service provides invoice management operations.
type Service struct {
	invoiceRepo repository.InvoiceRepository
	orderRepo   repository.OrderRepository
	config      *Config
	logger      logger.Logger

	// For invoice number generation
	seqMu      sync.Mutex
	seqMonth   string
	seqCounter int64
}

// NewService creates a new invoice service.
func NewService(
	invoiceRepo repository.InvoiceRepository,
	orderRepo repository.OrderRepository,
	log logger.Logger,
	config *Config,
) *Service {
	if config == nil {
		config = DefaultConfig()
	}
	return &Service{
		invoiceRepo: invoiceRepo,
		orderRepo:   orderRepo,
		config:      config,
		logger:      log,
	}
}

// Generate generates an invoice for an order.
func (s *Service) Generate(ctx context.Context, orderID int64) (*Invoice, error) {
	// Check if invoice already exists
	existing, err := s.invoiceRepo.GetByOrderID(ctx, orderID)
	if err == nil && existing != nil {
		return s.toInvoice(existing), nil
	}

	// Get order details
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		s.logger.Error("Failed to get order for invoice", logger.Err(err), logger.F("orderID", orderID))
		return nil, ErrOrderNotFound
	}

	// Generate invoice number
	invoiceNo := s.GenerateInvoiceNo()

	// Build content
	content := &Content{
		CompanyName:    s.config.CompanyName,
		CompanyAddress: s.config.CompanyAddress,
		TaxID:          s.config.TaxID,
		Items: []LineItem{
			{
				Description: fmt.Sprintf("Plan subscription - Order %s", order.OrderNo),
				Quantity:    1,
				UnitPrice:   order.OriginalAmount,
				Amount:      order.OriginalAmount,
			},
		},
		Subtotal: order.OriginalAmount,
		Discount: order.DiscountAmount,
		Total:    order.PayAmount,
	}

	// Get user info if available
	if order.User != nil {
		content.CustomerName = order.User.Username
		content.CustomerEmail = order.User.Email
	}

	contentJSON, err := json.Marshal(content)
	if err != nil {
		s.logger.Error("Failed to marshal invoice content", logger.Err(err))
		return nil, ErrInvalidInvoice
	}

	// Create invoice record
	repoInvoice := &repository.Invoice{
		InvoiceNo: invoiceNo,
		OrderID:   orderID,
		UserID:    order.UserID,
		Amount:    order.PayAmount,
		Content:   string(contentJSON),
	}

	if err := s.invoiceRepo.Create(ctx, repoInvoice); err != nil {
		s.logger.Error("Failed to create invoice", logger.Err(err))
		return nil, err
	}

	s.logger.Info("Invoice generated",
		logger.F("invoiceNo", invoiceNo),
		logger.F("orderID", orderID),
		logger.F("amount", order.PayAmount))

	return s.toInvoice(repoInvoice), nil
}

// GenerateInvoiceNo generates a unique invoice number.
func (s *Service) GenerateInvoiceNo() string {
	s.seqMu.Lock()
	defer s.seqMu.Unlock()

	currentMonth := time.Now().Format("200601")

	// Reset counter if month changed
	if s.seqMonth != currentMonth {
		s.seqMonth = currentMonth
		s.seqCounter = 0
	}

	s.seqCounter++
	return fmt.Sprintf(s.config.NumberFormat, currentMonth, s.seqCounter)
}

// GetByID retrieves an invoice by ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*Invoice, error) {
	repoInvoice, err := s.invoiceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}
	return s.toInvoice(repoInvoice), nil
}

// GetByInvoiceNo retrieves an invoice by invoice number.
func (s *Service) GetByInvoiceNo(ctx context.Context, invoiceNo string) (*Invoice, error) {
	repoInvoice, err := s.invoiceRepo.GetByInvoiceNo(ctx, invoiceNo)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}
	return s.toInvoice(repoInvoice), nil
}

// GetByOrder retrieves an invoice by order ID.
func (s *Service) GetByOrder(ctx context.Context, orderID int64) (*Invoice, error) {
	repoInvoice, err := s.invoiceRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}
	return s.toInvoice(repoInvoice), nil
}

// ListByUser lists invoices for a user with pagination.
func (s *Service) ListByUser(ctx context.Context, userID int64, page, pageSize int) ([]*Invoice, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	repoInvoices, total, err := s.invoiceRepo.ListByUser(ctx, userID, pageSize, offset)
	if err != nil {
		s.logger.Error("Failed to list invoices", logger.Err(err), logger.F("userID", userID))
		return nil, 0, err
	}

	invoices := make([]*Invoice, len(repoInvoices))
	for i, ri := range repoInvoices {
		invoices[i] = s.toInvoice(ri)
	}

	return invoices, total, nil
}

// List lists all invoices with filter and pagination.
func (s *Service) List(ctx context.Context, filter repository.InvoiceFilter, page, pageSize int) ([]*Invoice, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	repoInvoices, total, err := s.invoiceRepo.List(ctx, filter, pageSize, offset)
	if err != nil {
		s.logger.Error("Failed to list invoices", logger.Err(err))
		return nil, 0, err
	}

	invoices := make([]*Invoice, len(repoInvoices))
	for i, ri := range repoInvoices {
		invoices[i] = s.toInvoice(ri)
	}

	return invoices, total, nil
}


// GeneratePDF generates a PDF for an invoice.
// This is a basic implementation that returns a simple text-based PDF.
// In production, use a proper PDF library like gofpdf or wkhtmltopdf.
func (s *Service) GeneratePDF(ctx context.Context, invoice *Invoice) ([]byte, error) {
	if invoice == nil {
		return nil, ErrInvoiceNotFound
	}

	// Basic PDF content (placeholder implementation)
	// In production, use a proper PDF generation library
	content := fmt.Sprintf(`%%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>
endobj
4 0 obj
<< /Length 500 >>
stream
BT
/F1 24 Tf
50 750 Td
(INVOICE) Tj
/F1 12 Tf
0 -30 Td
(Invoice No: %s) Tj
0 -20 Td
(Date: %s) Tj
0 -40 Td
(From: %s) Tj
0 -20 Td
(%s) Tj
0 -40 Td
(Amount: %.2f) Tj
ET
endstream
endobj
5 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj
xref
0 6
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000266 00000 n 
0000000817 00000 n 
trailer
<< /Size 6 /Root 1 0 R >>
startxref
896
%%%%EOF`,
		invoice.InvoiceNo,
		invoice.CreatedAt,
		s.config.CompanyName,
		s.config.CompanyAddress,
		float64(invoice.Amount)/100,
	)

	return []byte(content), nil
}

// UpdatePDFPath updates the PDF path for an invoice.
func (s *Service) UpdatePDFPath(ctx context.Context, id int64, pdfPath string) error {
	invoice, err := s.invoiceRepo.GetByID(ctx, id)
	if err != nil {
		return ErrInvoiceNotFound
	}

	invoice.PDFPath = pdfPath
	return s.invoiceRepo.Update(ctx, invoice)
}

// GetConfig returns the invoice configuration.
func (s *Service) GetConfig() *Config {
	return s.config
}

// toInvoice converts a repository invoice to a service invoice.
func (s *Service) toInvoice(ri *repository.Invoice) *Invoice {
	invoice := &Invoice{
		ID:        ri.ID,
		InvoiceNo: ri.InvoiceNo,
		OrderID:   ri.OrderID,
		UserID:    ri.UserID,
		Amount:    ri.Amount,
		PDFPath:   ri.PDFPath,
		CreatedAt: ri.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// Parse content JSON
	if ri.Content != "" {
		var content Content
		if err := json.Unmarshal([]byte(ri.Content), &content); err == nil {
			invoice.Content = &content
		}
	}

	return invoice
}
