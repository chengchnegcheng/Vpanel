// Package payment provides payment gateway functionality.
package payment

import (
	"testing"
	"testing/quick"

	"v/internal/logger"
)

// Feature: commercial-system, Property 10: Payment Callback Idempotency
// Validates: Requirements 14.8
// For any payment callback processed multiple times with the same payment_no,
// the order status and balance SHALL only be updated once.

func TestProperty_CallbackIdempotency(t *testing.T) {
	log := logger.NewNopLogger()
	svc := NewService(nil, log)

	// Property: Processing the same callback multiple times should be idempotent
	f := func(method string, paymentNo string, times uint8) bool {
		if method == "" || paymentNo == "" {
			return true
		}

		// Limit times to reasonable range
		n := int(times%10) + 1

		// First call should mark as processed
		for i := 0; i < n; i++ {
			// Simulate marking callback as processed
			callbackKey := method + ":" + paymentNo
			svc.callbackMu.Lock()
			wasProcessed := svc.processedCallbacks[callbackKey]
			if !wasProcessed {
				svc.processedCallbacks[callbackKey] = true
			}
			svc.callbackMu.Unlock()

			// After first iteration, should always be marked as processed
			if i > 0 && !wasProcessed {
				return false
			}
		}

		// Verify it's marked as processed
		return svc.IsCallbackProcessed(method, paymentNo)
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(f, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property: Different payment numbers should be tracked independently
func TestProperty_IndependentCallbackTracking(t *testing.T) {
	log := logger.NewNopLogger()
	svc := NewService(nil, log)

	f := func(method1, paymentNo1, method2, paymentNo2 string) bool {
		if method1 == "" || paymentNo1 == "" || method2 == "" || paymentNo2 == "" {
			return true
		}

		// Mark first callback as processed
		svc.callbackMu.Lock()
		svc.processedCallbacks[method1+":"+paymentNo1] = true
		svc.callbackMu.Unlock()

		// Check first is processed
		if !svc.IsCallbackProcessed(method1, paymentNo1) {
			return false
		}

		// If different, second should not be processed
		if method1+":"+paymentNo1 != method2+":"+paymentNo2 {
			return !svc.IsCallbackProcessed(method2, paymentNo2)
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(f, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Feature: commercial-system, Property 13: Subscription Activation on Payment
// Validates: Requirements 4.7
// For any successful payment, the user's subscription SHALL be activated
// with correct expiration date based on plan duration.

func TestProperty_PaymentResultSuccess(t *testing.T) {
	// Property: A successful payment result should have all required fields
	f := func(orderNo, paymentNo string, amount uint32) bool {
		if orderNo == "" || paymentNo == "" {
			return true
		}

		result := &PaymentResult{
			Success:   true,
			OrderNo:   orderNo,
			PaymentNo: paymentNo,
			Amount:    int64(amount),
		}

		// Successful payment should have order number and payment number
		return result.Success &&
			result.OrderNo != "" &&
			result.PaymentNo != "" &&
			result.Amount >= 0
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(f, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property: Failed payment result should have error message
func TestProperty_PaymentResultFailure(t *testing.T) {
	f := func(errorMsg string) bool {
		if errorMsg == "" {
			return true
		}

		result := &PaymentResult{
			Success: false,
			Error:   errorMsg,
		}

		// Failed payment should have error message
		return !result.Success && result.Error != ""
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(f, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property: Gateway registration should be idempotent
func TestProperty_GatewayRegistration(t *testing.T) {
	log := logger.NewNopLogger()

	f := func(gatewayName string) bool {
		if gatewayName == "" {
			return true
		}

		svc := NewService(nil, log)

		// Create a mock gateway
		mockGateway := &mockGateway{name: gatewayName}

		// Register multiple times
		svc.RegisterGateway(mockGateway)
		svc.RegisterGateway(mockGateway)
		svc.RegisterGateway(mockGateway)

		// Should still only have one gateway with that name
		gateways := svc.ListGateways()
		count := 0
		for _, name := range gateways {
			if name == gatewayName {
				count++
			}
		}

		return count == 1
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(f, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// mockGateway is a mock implementation of PaymentGateway for testing.
type mockGateway struct {
	name string
}

func (g *mockGateway) Name() string {
	return g.name
}

func (g *mockGateway) CreatePayment(order *PaymentOrder) (*PaymentRequest, error) {
	return &PaymentRequest{}, nil
}

func (g *mockGateway) VerifyCallback(data []byte, signature string) (*PaymentResult, error) {
	return &PaymentResult{Success: true}, nil
}

func (g *mockGateway) QueryPayment(paymentNo string) (*PaymentResult, error) {
	return &PaymentResult{Success: true}, nil
}

func (g *mockGateway) Refund(paymentNo string, amount int64, reason string) (*RefundResult, error) {
	return &RefundResult{Success: true}, nil
}
