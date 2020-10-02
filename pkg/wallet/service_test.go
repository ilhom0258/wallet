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
