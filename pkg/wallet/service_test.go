package wallet

import (
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
