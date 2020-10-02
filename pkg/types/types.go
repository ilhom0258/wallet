package types

// Money is defined as minimal sum for money in cents, dirams...
type Money int64

// PaymentCategory is category of the payment
type PaymentCategory string

// PaymentStatus is status of the payment
type PaymentStatus string

// Statuses of Payments
const (
	PaymentStatusOk         PaymentStatus = "OK"
	PaymentStatusFail       PaymentStatus = "FAIL"
	PaymentStatusInProgress PaymentStatus = "INPROGRESS"
)

//Payment defines paymnet information
type Payment struct {
	ID        string
	AccountID int64
	Amount    Money
	Category  PaymentCategory
	Status    PaymentStatus
}

// Phone - phone number of the user
type Phone string

// Account defines account information of a user
type Account struct {
	ID      int64
	Phone   Phone
	Balance Money
}
