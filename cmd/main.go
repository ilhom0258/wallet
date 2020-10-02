package main

import (
	"fmt"

	"github.com/ilhom0258/wallet/pkg/wallet"
)

func main() {
	svc := &wallet.Service{}
	account, err := svc.RegisterAccount("+992501182129")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = svc.Deposit(account.ID, 10)
	if err != nil {
		switch err {
		case wallet.ErrAccountNotFound:
			fmt.Println("Аккаунт пользователя не найден")
		case wallet.ErrAmountMustBePositive:
			fmt.Println("Сумма должна быть положительной")
		}
		return
	}
	fmt.Println(account.Balance)

	paymnet, err := svc.Pay(account.ID, 10, "Food")
	if err != nil {
		switch err {
		case wallet.ErrAccountNotFound:
			fmt.Println("Аккаунт не найден")
		case wallet.ErrAmountMustBePositive:
			fmt.Println("Сумма должна быть положительной")
		case wallet.ErrNotEnoughBalance:
			fmt.Println("Недостаточно суммы на балансе")
		}
		return
	}
	fmt.Println(paymnet.Status)
}
