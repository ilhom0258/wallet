package wallet

import (
	"fmt"
	"testing"

	"github.com/ilhom0258/wallet/pkg/types"
)

func TestService_FindAccountByID_success(t *testing.T) {
	svr := &Service{}
	phone := types.Phone("+992501182129")
	account, err := svr.RegisterAccount(phone)
	if err != nil {
		switch err {
		case ErrPhoneRegistered:
			errorString := "Номер телефона уже зарегистрирован"
			t.Errorf(errorString, phone)
		}
		return
	}
	result, err := svr.FindAccountByID(account.ID)
	if err != nil {
		switch err {
		case ErrAccountNotFound:
			errorString := "Аккаунт с таким ID не найден"
			t.Errorf(errorString, phone)
		}
		return
	}
	t.Log(result)
}

func TestService_FindAccountByID_ErrAccountNotFound(t *testing.T) {
	svr := &Service{}
	_, err := svr.FindAccountByID(10)
	if err != nil {
		switch err {
		case ErrAccountNotFound:
			t.Log(err)
		}
	}
}

func TestService_Reject_ErrPaymentNotFound(t *testing.T) {
	svr := &Service{}
	_, err := svr.FindPaymentByID("10")
	if err != nil {
		switch err {
		case ErrPaymentNotFound:
			t.Log("Оплата не была найдена")
		}
		return
	}
}

func TestService_Reject_success(t *testing.T) {
	svr := &Service{}
	phone := types.Phone("+992501182129")
	account, err := svr.RegisterAccount(phone)
	if err != nil {
		switch err {
		case ErrPhoneRegistered:
			t.Errorf("Номер уже используется %v", phone)
		}
		return
	}
	err = svr.Deposit(account.ID, types.Money(666))
	if err != nil {
		switch err {
		case ErrAccountNotFound:
			t.Errorf("Аккаунт c ID = %v не найден", account.ID)

		case ErrAmountMustBePositive:
			t.Error("Сумма должна быть больше нуля")
		}
		return
	}
	payment, err := svr.Pay(account.ID, types.Money(12), types.PaymentCategory("Clothes"))
	if err != nil {
		switch err {
		case ErrAmountMustBePositive:
			t.Error("Cумма оплаты должна быть больше нуля")
		case ErrNotEnoughBalance:
			t.Error("Сумма на балансе карты недостаточна для оплаты")
		}
		return
	}
	_, err = svr.FindPaymentByID(payment.ID)
	if err != nil {
		switch err {
		case ErrPaymentNotFound:
			t.Errorf("Оплата с таким ID = %v не найден", payment.ID)
		}
		return
	}
	err = svr.Reject(payment.ID)
	if err != nil {
		switch err {
		case ErrPaymentNotFound:
			t.Error("Аккаунт не найден")
		}
	}
	t.Logf("Оплата с ID = %v отменена", payment.ID)
}

func TestService_Repeat_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)

	payment, err := s.Repeat(payments[0].ID)
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Logf("success payment = %v", payment)
}

func TestService_FavoritePayment_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	favoriteName := "mobile"
	payment := payments[0]
	_, err = s.FavoritePayment(payment.ID, favoriteName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestService_PayFromFavorite_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	payment := payments[0]
	favoriteName := "mobile"
	favorite, err := s.FavoritePayment(payment.ID, favoriteName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	_, err = s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestService_Export_success(t *testing.T) {
	s := newTestService()
	_, err := s.RegisterAccount("+0000000001")
	_, err = s.RegisterAccount("+0000000002")
	_, err = s.RegisterAccount("+0000000003")
	_, err = s.RegisterAccount("+0000000004")
	_, err = s.RegisterAccount("+0000000005")
	_, err = s.RegisterAccount("+0000000006")
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	favoriteName := "mobile"
	payment := payments[0]
	_, err = s.FavoritePayment(payment.ID, favoriteName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err != nil {
		t.Errorf("error %v", err)
		return
	}
	s.Export("data")
}

func TestService_Import_success(t *testing.T) {
	s := newTestService()
	s.Import("data")
}

// =========== Helper methods
type testService struct {
	*Service
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccount = testAccount{
	phone:   "+992000000001",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, error) {
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can't deposity account, error = %v", err)
	}
	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can't make paymnet, index = %v, error = %v", i, err)
		}
	}
	return account, payments, nil
}
