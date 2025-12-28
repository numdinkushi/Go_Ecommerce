package service

import (
	"go-ecommerce-app/pkg/external/flutterwave"
)

type BankService struct {
	flutterwaveClient *flutterwave.Client
}

type Bank struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type VerifyAccountResult struct {
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	BankCode      string `json:"bank_code"`
}

func NewBankService(flutterwaveClient *flutterwave.Client) *BankService {
	return &BankService{
		flutterwaveClient: flutterwaveClient,
	}
}

func (s *BankService) GetBanks() ([]Bank, error) {
	flutterwaveBanks, err := s.flutterwaveClient.GetBanks("NG")
	if err != nil {
		return nil, err
	}

	// Convert Flutterwave bank format to our format
	banks := make([]Bank, len(flutterwaveBanks))
	for i, bank := range flutterwaveBanks {
		banks[i] = Bank{
			Code: bank.Code,
			Name: bank.Name,
		}
	}

	return banks, nil
}

func (s *BankService) VerifyAccount(accountNumber, bankCode string) (*VerifyAccountResult, error) {
	result, err := s.flutterwaveClient.VerifyAccount(accountNumber, bankCode)
	if err != nil {
		return nil, err
	}

	return &VerifyAccountResult{
		AccountNumber: result.Data.AccountNumber,
		AccountName:   result.Data.AccountName,
		BankCode:      bankCode,
	}, nil
}
