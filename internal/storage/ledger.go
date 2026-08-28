package storage

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TransferFunds(ctx context.Context, db *pgxpool.Pool, fromAccount string, toAccount string, amount float64) error {

	//check if amount is greater than 0
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}

	//check if not same account
	if fromAccount == toAccount {
		return fmt.Errorf("sender and receiver must be different")
	}
	//-------------------------------------------------------------------------------

	//same transaction id for both users for one transaction
	txID := uuid.New().String()

	//avoiding deadlock
	firstLock, secondLock := fromAccount, toAccount
	if toAccount < fromAccount {
		firstLock, secondLock = toAccount, fromAccount

	}


	//initiating transaction(now we will work with tx not db)
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	//if any problem occurs before commit then undo everything
	defer tx.Rollback(ctx)

	//-------------------------------------------------------------------------------


	var dummy int

	//lock the first user's row using firstlock not accountid for avoiding deadlock
	err = tx.QueryRow(ctx, `SELECT 1 FROM accounts where id=$1 FOR UPDATE`, firstLock).Scan(&dummy)
	if err != nil {
		return fmt.Errorf("failed to lock first account: %w", err)
	}

	//lock the second user's row 
	err = tx.QueryRow(ctx, `SELECT 1 FROM accounts where id=$1 FOR UPDATE`, secondLock).Scan(&dummy)
	if err != nil {
		return fmt.Errorf("failed to lock second account: %w", err)
	}

	//-------------------------------------------------------------------------------

	var senderBalance float64
	err = tx.QueryRow(ctx, `SELECT cached_balance FROM accounts WHERE id = $1`, fromAccount).Scan(&senderBalance)
	if err != nil {
		return fmt.Errorf("read sender balance: %w", err)
	}

	//make go check if sender has enough balance for sending rather than postgre checking everytime which adds overhead and latency
	if senderBalance < amount {
		return fmt.Errorf("insufficient funds")
	}

	//-------------------------------------------------------------------------------
	//removing balance from sender's account
	_, err = tx.Exec(ctx, `UPDATE accounts SET cached_balance=cached_balance-$1 WHERE id=$2`, amount, fromAccount)
	if err != nil {
		return fmt.Errorf("deduct sender balance: %w", err)
	}

	//adding balance to reciever's account
	_, err = tx.Exec(ctx, `UPDATE accounts SET cached_balance=cached_balance+$1 WHERE id=$2`, amount, toAccount)
	if err != nil {
		return fmt.Errorf("add receiver balance: %w", err)
	}

	//-------------------------------------------------------------------------------


	//updating the source of truth -> our double-entry ledger 
	_, err = tx.Exec(ctx, `INSERT INTO ledger_entries(transaction_id,account_id,amount)
	VALUES($1,$2,-$3),VALUES($1,$4,$3)`, txID, fromAccount, amount, toAccount)
	if err != nil {
		return fmt.Errorf("write ledger: %w", err)
	}

	//making permanent changes
	return tx.Commit(ctx)

}
