package storage
import(
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

func TransferFunds(ctx context.Context, db *pgxpool.Pool,fromAccount string,toAccount string,amount float64) error{


	if amount<=0{
				return fmt.Errorf("amount must be greater than zero")
	}

	if fromAccount==toAccount{
				return fmt.Errorf("sender and receiver must be different")
	}
	
	txID:=uuid.New().String()

	firstLock,secondLock:=fromAccount,toAccount
	if toAccount<fromAccount{
			firstLock,secondLock:=toAccount,fromAccount

	}
	tx,err:=db.Begin(ctx)
	if err!=nil{
		return err
	}

	defer tx.Rollback(ctx)

	var dummy int
	err=tx.QueryRow(ctx,`SELECT 1 FROM accounts where id=$1 FOR UPDATE`,firstLock).Scan(&dummy)
	if err!=nil{
		return fmt.Errorf("failed to lock first account: %w", err)
	}
	err=tx.QueryRow(ctx,`SELECT 1 FROM accounts where id=$1 FOR UPDATE`,secondLock).Scan(&dummy)
	if err!=nil{
		return fmt.Errorf("failed to lock second account: %w", err)
	}

	var senderBalance float64
	err = tx.QueryRow(ctx, `SELECT cached_balance FROM accounts WHERE id = $1`, fromAccount).Scan(&senderBalance)

	if senderBalance<amount{
		return fmt.Errorf("insufficient funds") 
	}

	_, err=tx.Exec(ctx,`UPDATE accounts SET cached_balance=cached_balance-$1 WHERE id=$2`,amount,fromAccount)
	if err != nil { return err }
	_, err=tx.Exec(ctx,`UPDATE accounts SET cached_balance=cached_balance+$1 WHERE id=$2`,amount,toAccount)
	if err != nil { return err }

	_,err=tx.Exec(ctx,`INSERT INTO ledger_entries(transaction_id,account_id,amount)
	VALUES($1,$2,-$3),VALUES($1,$4,$3)`,txID,fromAccount,amount,toAccount)
	if err!=nil {
		return err
	}
	return tx.Commit(ctx)

}
