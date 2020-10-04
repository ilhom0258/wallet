package wallet

import (
	"errors"

	"github.com/google/uuid"
	"github.com/ilhom0258/wallet/pkg/types"
)

//Service structure for wallet and other services
type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

//Error message
type Error string

func (e Error) Error() string {
	return string(e)
}

// ErrAccountNotFound = Common Error
var ErrAccountNotFound = errors.New("account not found")

// ErrAmountMustBePositive Common Error
var ErrAmountMustBePositive = errors.New("amount must be greater than zero")

// ErrPhoneRegistered Common Error
var ErrPhoneRegistered = errors.New("phone already registered")

//ErrNotEnoughBalance Common Error
var ErrNotEnoughBalance = errors.New("not enough balance in account")

//ErrPaymentNotFound Common Error
var ErrPaymentNotFound = errors.New("paymnet not found")

//ErrFavoriteNotFound Common Error
var ErrFavoriteNotFound = errors.New("favorite not found")

// RegisterAccount function for registering wallet account for user
func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}
	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)
	return account, nil
}

//Deposit function for
func (s *Service) Deposit(accountID int64, amount types.Money) error {
	var acc *types.Account
	if amount <= 0 {
		return ErrAmountMustBePositive
	}
	for _, account := range s.accounts {
		if account.ID == accountID {
			acc = account
		}
	}
	if acc == nil {
		return ErrAccountNotFound
	}
	acc.Balance += amount
	return nil
}

// Pay function for making payments
func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}
	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	account.Balance -= amount
	s.payments = append(s.payments, payment)
	return payment, nil
}

//FindAccountByID function that finds account by ID
func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

//FindPaymentByID function seaches for payment with ID
func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	var payment *types.Payment
	for _, pmnt := range s.payments {
		if paymentID == pmnt.ID && pmnt.Status == types.PaymentStatusInProgress {
			payment = pmnt
			break
		}
	}
	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	return payment, nil
}

//Reject cancels payment which ProgressStatus
//930777607
func (s *Service) Reject(paymentID string) error {
	var payment *types.Payment
	for _, pmnt := range s.payments {
		if pmnt.ID == paymentID {
			payment = pmnt
			break
		}
	}
	if payment == nil {
		return ErrPaymentNotFound
	}
	payment.Status = types.PaymentStatusFail
	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == payment.AccountID {
			account = acc
			break
		}
	}
	account.Balance += payment.Amount
	return nil
}

//Repeat function that repeats payment with different UUID
func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	var payment *types.Payment
	for _, pmnt := range s.payments {
		if pmnt.ID == paymentID {
			payment = pmnt
			break
		}
	}
	if payment == nil {
		return nil, ErrPaymentNotFound
	}
	payment, err := s.Pay(payment.AccountID, payment.Amount, payment.Category)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

//FavoritePayment function for creating favorite payment
func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {

	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}
	favorite := &types.Favorite{
		ID:        paymentID,
		Amount:    payment.Amount,
		AccountID: payment.AccountID,
		Name:      name,
		Category:  payment.Category,
	}
	s.favorites = append(s.favorites, favorite)
	return favorite, nil
}

//PayFromFavorite function for favorite payment for user
func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	var favorite *types.Favorite
	for _, fvrt := range s.favorites {
		if fvrt.ID == favoriteID {
			favorite = fvrt
			break
		}
	}
	if favorite == nil {
		return nil, ErrFavoriteNotFound
	}
	payment, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
	if err != nil {
		return nil, err
	}
	return payment, nil
}
