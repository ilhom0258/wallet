package wallet

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

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

//ErrWorkingDirectoryNotFound Common Error
var ErrWorkingDirectoryNotFound = errors.New("wd not found")

//ErrInParsing Common Error
var ErrInParsing = errors.New("error in parsing")

//ErrInReadingFromFile Common Error
var ErrInReadingFromFile = errors.New("error in reading from file")

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
	if payment == nil {
		return nil, ErrPaymentNotFound
	}
	favorite := &types.Favorite{
		ID:        uuid.New().String(),
		AccountID: payment.AccountID,
		Name:      name,
		Amount:    payment.Amount,
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

// ExportToFile exports data to file
func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Println(err)
		return ErrWorkingDirectoryNotFound
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()
	for _, account := range s.accounts {
		id := strconv.FormatInt(int64(account.ID), 10)
		balance := strconv.FormatInt(int64(account.Balance), 10)
		phone := string(account.Phone)
		_, err := file.Write([]byte(id + ";" + phone + ";" + balance + "|"))
		if err != nil {
			log.Print(err)
			return ErrWorkingDirectoryNotFound
		}
	}
	return nil
}

// ImportFromFile imports data from file
func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)
		return ErrWorkingDirectoryNotFound
	}
	content := make([]byte, 0)
	buf := make([]byte, 4)
	for {
		read, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		content = append(content, buf[:read]...)
	}
	data := string(content)
	temp := strings.Split(data, "|")
	for _, acc := range temp {
		accountData := strings.Split(acc, ";")
		if acc == "" {
			break
		}
		id, err := strconv.ParseInt(accountData[0], 10, 64)
		if err != nil {
			log.Print(err)
			return ErrInParsing
		}
		phone := accountData[1]
		balance, err := strconv.ParseInt(accountData[2], 10, 64)
		if err != nil {
			log.Print(err)
			return ErrInParsing
		}
		account := &types.Account{
			ID:      id,
			Phone:   types.Phone(phone),
			Balance: types.Money(balance),
		}
		s.nextAccountID = id
		s.accounts = append(s.accounts, account)
	}
	return nil
}

// Export - exports data to file
func (s *Service) Export(dir string) (err error) {
	accounts := s.accounts
	payments := s.payments
	favorites := s.favorites
	if len(accounts) != 0 {
		path, err := pathMaker(dir, "accounts.dump")
		err = exportAccounts(accounts, path)
		log.Printf("%v error in accounts", err)
	}
	if len(payments) != 0 {
		path, err := pathMaker(dir, "payments.dump")
		err = exportPayments(payments, path)
		log.Printf("%v error in paymentts", err)
	}
	if len(favorites) != 0 {
		path, err := pathMaker(dir, "favorites.dump")
		err = exportFavorites(favorites, path)
		log.Printf("%v error in favorites", err)
	}
	return nil
}

//Import - import data from file
func (s *Service) Import(dir string) (err error) {

	path, err := filepath.Abs(dir)
	fmt.Println(path)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	accountPath := path + "/accounts.dump"
	paymentPath := path + "/payments.dump"
	favoritePath := path + "/favorites.dump"
	if s.fileExist(accountPath) {
		log.Printf("here acc %v", accountPath)
		err = importAccounts(accountPath, s)
	}
	if s.fileExist(paymentPath) {
		log.Printf("here pay %v", paymentPath)
		err = importPayments(paymentPath, s)
	}
	if s.fileExist(favoritePath) {
		log.Printf("here fav %v", favoritePath)
		err = importFavorites(favoritePath, s)
	}
	return nil
}

//ExportAccountHistory takes an accountID and returns all payments
func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {

	var payments []types.Payment
	var account *types.Account

	for _, acc := range s.accounts {
		if accountID == acc.ID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	for _, payment := range s.payments {
		if accountID == payment.AccountID {
			payments = append(payments, *payment)
		}
	}
	return payments, nil
}

//HistoryToFiles exports account history into files
func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {

	var pmnts []*types.Payment
	for _, payment := range payments {
		pmnts = append(pmnts, &payment)
	}
	if len(payments) <= records && len(payments) > 0 {
		path, err := pathMaker(dir, "payments.dump")
		if err != nil {
			return err
		}
		err = exportPayments(pmnts, path)
	}
	if len(payments) > records {
		paymentsToExport := []*types.Payment{}
		paymentsCnt := 1
		for i, payment := range pmnts {
			fileName := fmt.Sprintf("payments%v.dump", paymentsCnt)
			path, err := pathMaker(dir, fileName)
			paymentsToExport = append(paymentsToExport, payment)
			if (i+1)%records == 0 && i != 0 {
				err = exportPayments(paymentsToExport, path)
				if err != nil {
					return err
				}
				paymentsCnt++
				paymentsToExport = []*types.Payment{}
			}
			if i+1 == len(pmnts) && (i+1)%records != 0 {
				err = exportPayments(paymentsToExport, path)
				if err != nil {
					return err
				}
			}

		}
	}
	return nil
}

//SumPayments calculates sum of payments amount with goroutines
func (s *Service) SumPayments(goroutines int) types.Money {
	if goroutines <= 1 || len(s.payments) == 1 {
		return regularSum(s.payments)
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	total := types.Money(0)

	index := 0
	prop := int(math.Floor(float64(len(s.payments)) / float64(goroutines)))
	data := make([][]*types.Payment, prop+1)
	log.Print(s.payments)
	for i := 0; i < len(s.payments); i+=goroutines {
		end := i + goroutines
		if end > len(s.payments){
			end = len(s.payments)
		}
		data[index] = s.payments[i:end]
		index++
	}
	for _, pntSlice := range data {
		wg.Add(1)
		log.Print(pntSlice)
		go concurrentSum(&total, pntSlice, &wg, &mu)
	}
	wg.Wait()
	return total
}

// Helpers

func regularSum(payments []*types.Payment) types.Money {
	sum := types.Money(0)
	for _, payment := range payments {
		sum += payment.Amount
	}
	return sum
}

func concurrentSum(total *types.Money, payments []*types.Payment, wg *sync.WaitGroup, mu *sync.Mutex) {
	sum := types.Money(0)
	mu.Lock()
	for _, payment := range payments {
		sum += payment.Amount
	}
	*total += sum
	mu.Unlock()
	wg.Done()
}

func exportAccounts(accounts []*types.Account, dir string) (err error) {
	data := ""
	for _, account := range accounts {
		id := strconv.FormatInt(int64(account.ID), 10)
		balance := strconv.FormatInt(int64(account.Balance), 10)
		phone := string(account.Phone)
		data += id + ";" + phone + ";" + balance + "\n"
	}
	err = ioutil.WriteFile(dir, []byte(data), 0777)
	if err != nil {
		log.Print(err)
		return ErrWorkingDirectoryNotFound
	}
	return nil
}

func exportPayments(payments []*types.Payment, dir string) (err error) {
	data := ""
	for _, payment := range payments {
		id := string(payment.ID)
		accID := strconv.FormatInt(int64(payment.AccountID), 10)
		amount := strconv.FormatInt(int64(payment.Amount), 10)
		cat := string(payment.Category)
		stat := string(payment.Status)
		data += id + ";" + accID + ";" + amount + ";" + cat + ";" + stat + "\n"
	}
	err = ioutil.WriteFile(dir, []byte(data), 0777)
	if err != nil {
		log.Print(err)
		return ErrWorkingDirectoryNotFound
	}
	return nil
}

func exportFavorites(favorites []*types.Favorite, dir string) (err error) {
	data := ""
	for _, favorite := range favorites {
		id := string(favorite.ID)
		accID := strconv.FormatInt(int64(favorite.AccountID), 10)
		amount := strconv.FormatInt(int64(favorite.Amount), 10)
		cat := string(favorite.Category)
		name := string(favorite.Name)
		data += id + ";" + accID + ";" + name + ";" + amount + ";" + cat + "\n"
	}
	err = ioutil.WriteFile(dir, []byte(data), 0777)
	if err != nil {
		log.Print(err)
		return ErrWorkingDirectoryNotFound
	}
	return nil
}

func importAccounts(path string, s *Service) error {
	dataRaw, err := s.getDataFromFile(path)
	if err != nil {
		return ErrWorkingDirectoryNotFound
	}
	data := string(dataRaw)
	accounts, err := parseAccounts(data)
	for _, account := range accounts {
		if !isAccountInService(account, s) {
			s.accounts = append(s.accounts, account)
			if account.ID >= s.nextAccountID {
				s.nextAccountID = account.ID + 1
			}
		}
	}
	return nil
}

func importPayments(path string, s *Service) error {
	dataRaw, err := s.getDataFromFile(path)
	if err != nil {
		return ErrWorkingDirectoryNotFound
	}
	data := string(dataRaw)
	payments, err := parsePayments(data)
	for _, payment := range payments {
		if !isPaymentInService(payment, s) {
			s.payments = append(s.payments, payment)
		}
	}
	return nil
}

func importFavorites(path string, s *Service) error {
	dataRaw, err := s.getDataFromFile(path)
	if err != nil {
		return ErrWorkingDirectoryNotFound
	}
	data := string(dataRaw)
	favorites, err := parseFavorites(data)
	for _, favorite := range favorites {
		if !isFavoriteInService(favorite, s) {
			s.favorites = append(s.favorites, favorite)
		}
	}
	return nil
}

func isAccountInService(info *types.Account, s *Service) bool {
	for _, account := range s.accounts {
		if reflect.DeepEqual(account, info) {
			return true
		}
	}
	return false
}

func isPaymentInService(info *types.Payment, s *Service) bool {
	for _, payment := range s.payments {
		if reflect.DeepEqual(payment, info) {
			return true
		}
	}
	return false
}

func isFavoriteInService(info *types.Favorite, s *Service) bool {
	for _, favorite := range s.favorites {
		if reflect.DeepEqual(favorite, info) {
			return true
		}
	}
	return false
}

func parseAccounts(data string) ([]*types.Account, error) {
	var accounts []*types.Account
	dataRaw := strings.Split(data, "\n")
	for _, item := range dataRaw {
		info := strings.Split(item, ";")
		if len(strings.Trim(item, " ")) == 0 {
			break
		}
		ID, err := strconv.ParseInt(info[0], 10, 64)
		if err != nil {
			return nil, ErrInParsing
		}
		phone := types.Phone(info[1])
		balance, err := strconv.ParseInt(info[2], 10, 64)
		if err != nil {
			return nil, ErrInParsing
		}
		account := &types.Account{
			ID:      ID,
			Phone:   phone,
			Balance: types.Money(balance),
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func parsePayments(data string) ([]*types.Payment, error) {
	var payments []*types.Payment
	dataRaw := strings.Split(data, "\n")
	for _, item := range dataRaw {
		info := strings.Split(item, ";")
		if len(strings.Trim(item, " ")) == 0 {
			break
		}
		ID := string(info[0])
		accountID, err := strconv.ParseInt(info[1], 10, 64)
		if err != nil {
			return nil, ErrInParsing
		}
		amount, err := strconv.ParseInt(info[2], 10, 64)
		if err != nil {
			return nil, ErrInParsing
		}
		category := string(info[3])
		status := types.PaymentStatus(info[4])
		payment := &types.Payment{
			ID:        ID,
			AccountID: accountID,
			Amount:    types.Money(amount),
			Category:  types.PaymentCategory(category),
			Status:    status,
		}
		payments = append(payments, payment)
	}
	return payments, nil
}

func parseFavorites(data string) ([]*types.Favorite, error) {
	var favorites []*types.Favorite
	dataRaw := strings.Split(data, "\n")
	for _, item := range dataRaw {
		info := strings.Split(item, ";")
		if len(strings.Trim(item, " ")) == 0 {
			break
		}
		ID := string(info[0])
		accountID, err := strconv.ParseInt(info[1], 10, 64)
		if err != nil {
			return nil, ErrInParsing
		}
		name := string(info[2])
		amount, err := strconv.ParseInt(info[3], 10, 64)
		if err != nil {
			return nil, ErrInParsing
		}
		category := string(info[4])
		favorite := &types.Favorite{
			ID:        ID,
			AccountID: accountID,
			Name:      name,
			Amount:    types.Money(amount),
			Category:  types.PaymentCategory(category),
		}
		favorites = append(favorites, favorite)
	}
	return favorites, nil
}

func pathMaker(dir string, fileName string) (string, error) {
	path, err := filepath.Abs(dir)
	if err != nil {
		log.Print(err)
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0777)
	}
	return path + "/" + fileName, nil
}

func (s *Service) fileExist(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func (s *Service) getDataFromFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Println(err)
		}
	}()

	content := make([]byte, 0)
	buf := make([]byte, 4096)

	for {
		read, err := file.Read(buf)
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Println(err)
			return "", err
		}

		content = append(content, buf[:read]...)
	}

	data := string(content)
	return data, nil
}
