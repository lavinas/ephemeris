package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	conn *sql.DB
}

func Connect(dsn string) (*DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com postgres: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("erro ao conectar no postgres (%s): %w", dsn, err)
	}

	return &DB{conn: conn}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

// RunMigrations executa o script SQL de migração se fornecido
func (d *DB) RunMigrations(migrationFile string) error {
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("arquivo de migração não encontrado em %s: %w", migrationFile, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := d.conn.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("erro ao executar migração %s: %w", migrationFile, err)
	}

	return nil
}

// UpsertItem insere ou atualiza um item na base
func (d *DB) UpsertItem(ctx context.Context, item *DBItem) error {
	query := `
		INSERT INTO pluggy_item (
			id, connector_id, connector_name, connector_image_url, connector_primary_color,
			status, execution_status, client_user_id, error_code, error_message,
			last_updated_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, COALESCE($12, NOW()), NOW())
		ON CONFLICT (id) DO UPDATE SET
			connector_id = EXCLUDED.connector_id,
			connector_name = EXCLUDED.connector_name,
			connector_image_url = EXCLUDED.connector_image_url,
			connector_primary_color = EXCLUDED.connector_primary_color,
			status = EXCLUDED.status,
			execution_status = EXCLUDED.execution_status,
			client_user_id = EXCLUDED.client_user_id,
			error_code = EXCLUDED.error_code,
			error_message = EXCLUDED.error_message,
			last_updated_at = EXCLUDED.last_updated_at,
			updated_at = NOW();
	`
	_, err := d.conn.ExecContext(ctx, query,
		item.ID, item.ConnectorID, item.ConnectorName, item.ConnectorImageURL, item.ConnectorPrimaryColor,
		item.Status, item.ExecutionStatus, item.ClientUserID, item.ErrorCode, item.ErrorMessage,
		item.LastUpdatedAt, item.CreatedAt,
	)
	return err
}

// DeleteItem remove um item e todos os dados associados via cascade
func (d *DB) DeleteItem(ctx context.Context, id string) error {
	query := `DELETE FROM pluggy_item WHERE id = $1`
	_, err := d.conn.ExecContext(ctx, query, id)
	return err
}

// UpsertAccount insere ou atualiza uma conta bancária ou cartão
func (d *DB) UpsertAccount(ctx context.Context, acc *DBAccount) error {
	query := `
		INSERT INTO pluggy_account (
			id, item_id, type, subtype, name, marketing_name, balance, currency_code, number, owner, tax_number,
			bank_agency, bank_transfer_number, bank_closing_balance, bank_overdraft_contracted_limit,
			bank_overdraft_used_limit, bank_automatically_invested_balance, bank_has_reserved_balance,
			credit_brand, credit_level, credit_limit, credit_available_limit,
			credit_balance_close_date, credit_balance_due_date, credit_minimum_payment,
			credit_status, credit_holder_type, raw_data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24, $25, $26, $27, $28::jsonb, NOW(), NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			item_id = EXCLUDED.item_id,
			type = EXCLUDED.type,
			subtype = EXCLUDED.subtype,
			name = EXCLUDED.name,
			marketing_name = EXCLUDED.marketing_name,
			balance = EXCLUDED.balance,
			currency_code = EXCLUDED.currency_code,
			number = EXCLUDED.number,
			owner = EXCLUDED.owner,
			tax_number = EXCLUDED.tax_number,
			bank_agency = EXCLUDED.bank_agency,
			bank_transfer_number = EXCLUDED.bank_transfer_number,
			bank_closing_balance = EXCLUDED.bank_closing_balance,
			bank_overdraft_contracted_limit = EXCLUDED.bank_overdraft_contracted_limit,
			bank_overdraft_used_limit = EXCLUDED.bank_overdraft_used_limit,
			bank_automatically_invested_balance = EXCLUDED.bank_automatically_invested_balance,
			bank_has_reserved_balance = EXCLUDED.bank_has_reserved_balance,
			credit_brand = EXCLUDED.credit_brand,
			credit_level = EXCLUDED.credit_level,
			credit_limit = EXCLUDED.credit_limit,
			credit_available_limit = EXCLUDED.credit_available_limit,
			credit_balance_close_date = EXCLUDED.credit_balance_close_date,
			credit_balance_due_date = EXCLUDED.credit_balance_due_date,
			credit_minimum_payment = EXCLUDED.credit_minimum_payment,
			credit_status = EXCLUDED.credit_status,
			credit_holder_type = EXCLUDED.credit_holder_type,
			raw_data = EXCLUDED.raw_data,
			updated_at = NOW();
	`
	rawData := acc.RawData
	if rawData == "" {
		rawData = "{}"
	}

	_, err := d.conn.ExecContext(ctx, query,
		acc.ID, acc.ItemID, acc.Type, acc.Subtype, acc.Name, acc.MarketingName, acc.Balance, acc.CurrencyCode, acc.Number, acc.Owner, acc.TaxNumber,
		acc.BankAgency, acc.BankTransferNumber, acc.BankClosingBalance, acc.BankOverdraftContractedLimit,
		acc.BankOverdraftUsedLimit, acc.BankAutomaticallyInvestedBalance, acc.BankHasReservedBalance,
		acc.CreditBrand, acc.CreditLevel, acc.CreditLimit, acc.CreditAvailableLimit,
		acc.CreditBalanceCloseDate, acc.CreditBalanceDueDate, acc.CreditMinimumPayment,
		acc.CreditStatus, acc.CreditHolderType, rawData,
	)
	return err
}

// UpsertTransaction insere ou atualiza uma transação bancária ou do cartão
func (d *DB) UpsertTransaction(ctx context.Context, tx *DBTransaction) error {
	query := `
		INSERT INTO pluggy_transaction (
			id, account_id, item_id, description, description_raw, amount, currency_code, date,
			category, category_id, type, status, balance, provider_code, merchant_name, merchant_cnpj,
			payment_data, raw_data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17::jsonb, $18::jsonb, NOW(), NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			account_id = EXCLUDED.account_id,
			item_id = EXCLUDED.item_id,
			description = EXCLUDED.description,
			description_raw = EXCLUDED.description_raw,
			amount = EXCLUDED.amount,
			currency_code = EXCLUDED.currency_code,
			date = EXCLUDED.date,
			category = EXCLUDED.category,
			category_id = EXCLUDED.category_id,
			type = EXCLUDED.type,
			status = EXCLUDED.status,
			balance = EXCLUDED.balance,
			provider_code = EXCLUDED.provider_code,
			merchant_name = EXCLUDED.merchant_name,
			merchant_cnpj = EXCLUDED.merchant_cnpj,
			payment_data = EXCLUDED.payment_data,
			raw_data = EXCLUDED.raw_data,
			updated_at = NOW();
	`
	paymentData := tx.PaymentData
	if paymentData == "" {
		paymentData = "{}"
	}
	rawData := tx.RawData
	if rawData == "" {
		rawData = "{}"
	}

	_, err := d.conn.ExecContext(ctx, query,
		tx.ID, tx.AccountID, tx.ItemID, tx.Description, tx.DescriptionRaw, tx.Amount, tx.CurrencyCode, tx.Date,
		tx.Category, tx.CategoryID, tx.Type, tx.Status, tx.Balance, tx.ProviderCode, tx.MerchantName, tx.MerchantCNPJ,
		paymentData, rawData,
	)
	return err
}

// UpsertCreditCardBill insere ou atualiza faturas de cartão de crédito
func (d *DB) UpsertCreditCardBill(ctx context.Context, bill *DBCreditCardBill) error {
	query := `
		INSERT INTO pluggy_credit_card_bill (
			id, account_id, due_date, bill_closing_date, total_amount, currency_code,
			minimum_payment_amount, allows_installments, finance_charges, payments, raw_data,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10::jsonb, $11::jsonb, NOW(), NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			account_id = EXCLUDED.account_id,
			due_date = EXCLUDED.due_date,
			bill_closing_date = EXCLUDED.bill_closing_date,
			total_amount = EXCLUDED.total_amount,
			currency_code = EXCLUDED.currency_code,
			minimum_payment_amount = EXCLUDED.minimum_payment_amount,
			allows_installments = EXCLUDED.allows_installments,
			finance_charges = EXCLUDED.finance_charges,
			payments = EXCLUDED.payments,
			raw_data = EXCLUDED.raw_data,
			updated_at = NOW();
	`
	financeCharges := bill.FinanceCharges
	if financeCharges == "" {
		financeCharges = "[]"
	}
	payments := bill.Payments
	if payments == "" {
		payments = "[]"
	}
	rawData := bill.RawData
	if rawData == "" {
		rawData = "{}"
	}

	_, err := d.conn.ExecContext(ctx, query,
		bill.ID, bill.AccountID, bill.DueDate, bill.BillClosingDate, bill.TotalAmount, bill.CurrencyCode,
		bill.MinimumPaymentAmount, bill.AllowsInstallments, financeCharges, payments, rawData,
	)
	return err
}

// UpsertInvestment insere ou atualiza uma aplicação financeira
func (d *DB) UpsertInvestment(ctx context.Context, inv *DBInvestment) error {
	query := `
		INSERT INTO pluggy_investment (
			id, item_id, name, code, isin, type, subtype, balance, amount, currency_code,
			due_date, rate, rate_type, fixed_annual_rate, last_month_rate, annual_rate,
			status, institution, raw_data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19::jsonb, NOW(), NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			item_id = EXCLUDED.item_id,
			name = EXCLUDED.name,
			code = EXCLUDED.code,
			isin = EXCLUDED.isin,
			type = EXCLUDED.type,
			subtype = EXCLUDED.subtype,
			balance = EXCLUDED.balance,
			amount = EXCLUDED.amount,
			currency_code = EXCLUDED.currency_code,
			due_date = EXCLUDED.due_date,
			rate = EXCLUDED.rate,
			rate_type = EXCLUDED.rate_type,
			fixed_annual_rate = EXCLUDED.fixed_annual_rate,
			last_month_rate = EXCLUDED.last_month_rate,
			annual_rate = EXCLUDED.annual_rate,
			status = EXCLUDED.status,
			institution = EXCLUDED.institution,
			raw_data = EXCLUDED.raw_data,
			updated_at = NOW();
	`
	rawData := inv.RawData
	if rawData == "" {
		rawData = "{}"
	}

	_, err := d.conn.ExecContext(ctx, query,
		inv.ID, inv.ItemID, inv.Name, inv.Code, inv.ISIN, inv.Type, inv.Subtype, inv.Balance, inv.Amount, inv.CurrencyCode,
		inv.DueDate, inv.Rate, inv.RateType, inv.FixedAnnualRate, inv.LastMonthRate, inv.AnnualRate,
		inv.Status, inv.Institution, rawData,
	)
	return err
}

// UpsertInvestmentTransaction insere ou atualiza movimentações de investimentos
func (d *DB) UpsertInvestmentTransaction(ctx context.Context, tx *DBInvestmentTransaction) error {
	query := `
		INSERT INTO pluggy_investment_transaction (
			id, investment_id, trade_date, date, type, movement_type, description,
			quantity, value, amount, net_amount, agreed_rate, brokerage_number,
			expenses, raw_data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14::jsonb, $15::jsonb, NOW(), NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			investment_id = EXCLUDED.investment_id,
			trade_date = EXCLUDED.trade_date,
			date = EXCLUDED.date,
			type = EXCLUDED.type,
			movement_type = EXCLUDED.movement_type,
			description = EXCLUDED.description,
			quantity = EXCLUDED.quantity,
			value = EXCLUDED.value,
			amount = EXCLUDED.amount,
			net_amount = EXCLUDED.net_amount,
			agreed_rate = EXCLUDED.agreed_rate,
			brokerage_number = EXCLUDED.brokerage_number,
			expenses = EXCLUDED.expenses,
			raw_data = EXCLUDED.raw_data,
			updated_at = NOW();
	`
	expenses := tx.Expenses
	if expenses == "" {
		expenses = "{}"
	}
	rawData := tx.RawData
	if rawData == "" {
		rawData = "{}"
	}

	_, err := d.conn.ExecContext(ctx, query,
		tx.ID, tx.InvestmentID, tx.TradeDate, tx.Date, tx.Type, tx.MovementType, tx.Description,
		tx.Quantity, tx.Value, tx.Amount, tx.NetAmount, tx.AgreedRate, tx.BrokerageNumber,
		expenses, rawData,
	)
	return err
}

// SaveWebhookEvent armazena evento bruto recebido no webhook
func (d *DB) SaveWebhookEvent(ctx context.Context, event, eventID, itemID, accountID, triggeredBy string, payload any, status, errMsg string) error {
	payloadBytes, _ := json.Marshal(payload)
	query := `
		INSERT INTO pluggy_webhook_event (
			event, event_id, item_id, account_id, triggered_by, payload, status, error_message, received_at
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, NOW())
	`
	_, err := d.conn.ExecContext(ctx, query, event, eventID, itemID, accountID, triggeredBy, string(payloadBytes), status, errMsg)
	return err
}

// GetPortfolioSummary consulta os totais na view vw_portfolio_summary
func (d *DB) GetPortfolioSummary(ctx context.Context) (*PortfolioSummary, error) {
	query := `
		SELECT 
			total_bank_balance, total_credit_spent, total_investments_balance,
			total_connected_items, total_bank_accounts, total_credit_cards, total_investments
		FROM vw_portfolio_summary;
	`
	var s PortfolioSummary
	err := d.conn.QueryRowContext(ctx, query).Scan(
		&s.TotalBankBalance, &s.TotalCreditSpent, &s.TotalInvestmentsBalance,
		&s.TotalConnectedItems, &s.TotalBankAccounts, &s.TotalCreditCards, &s.TotalInvestments,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListItems retorna todas as conexões cadastradas
func (d *DB) ListItems(ctx context.Context) ([]DBItem, error) {
	query := `
		SELECT 
			i.id, i.connector_id, i.connector_name, i.connector_image_url, i.connector_primary_color,
			i.status, i.execution_status, i.client_user_id, COALESCE(i.error_code, ''), COALESCE(i.error_message, ''),
			i.last_updated_at, i.created_at, i.updated_at,
			(SELECT COUNT(*) FROM pluggy_account WHERE item_id = i.id) as accounts_count,
			(SELECT COUNT(*) FROM pluggy_investment WHERE item_id = i.id) as investments_count,
			(SELECT COUNT(*) FROM pluggy_transaction WHERE item_id = i.id) as transactions_count
		FROM pluggy_item i
		ORDER BY i.created_at DESC;
	`
	rows, err := d.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DBItem
	for rows.Next() {
		var item DBItem
		err := rows.Scan(
			&item.ID, &item.ConnectorID, &item.ConnectorName, &item.ConnectorImageURL, &item.ConnectorPrimaryColor,
			&item.Status, &item.ExecutionStatus, &item.ClientUserID, &item.ErrorCode, &item.ErrorMessage,
			&item.LastUpdatedAt, &item.CreatedAt, &item.UpdatedAt,
			&item.AccountsCount, &item.InvestmentsCount, &item.TransactionsCount,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// GetItem busca uma única conexão
func (d *DB) GetItem(ctx context.Context, id string) (*DBItem, error) {
	query := `
		SELECT 
			i.id, i.connector_id, i.connector_name, i.connector_image_url, i.connector_primary_color,
			i.status, i.execution_status, i.client_user_id, COALESCE(i.error_code, ''), COALESCE(i.error_message, ''),
			i.last_updated_at, i.created_at, i.updated_at,
			(SELECT COUNT(*) FROM pluggy_account WHERE item_id = i.id) as accounts_count,
			(SELECT COUNT(*) FROM pluggy_investment WHERE item_id = i.id) as investments_count,
			(SELECT COUNT(*) FROM pluggy_transaction WHERE item_id = i.id) as transactions_count
		FROM pluggy_item i
		WHERE i.id = $1;
	`
	var item DBItem
	err := d.conn.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.ConnectorID, &item.ConnectorName, &item.ConnectorImageURL, &item.ConnectorPrimaryColor,
		&item.Status, &item.ExecutionStatus, &item.ClientUserID, &item.ErrorCode, &item.ErrorMessage,
		&item.LastUpdatedAt, &item.CreatedAt, &item.UpdatedAt,
		&item.AccountsCount, &item.InvestmentsCount, &item.TransactionsCount,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ListAccounts retorna contas filtradas por tipo (BANK ou CREDIT)
func (d *DB) ListAccounts(ctx context.Context, accountType string) ([]DBAccount, error) {
	query := `
		SELECT 
			a.id, a.item_id, a.type, a.subtype, a.name, COALESCE(a.marketing_name, ''),
			a.balance, a.currency_code, COALESCE(a.number, ''), COALESCE(a.owner, ''), COALESCE(a.tax_number, ''),
			COALESCE(a.bank_agency, ''), COALESCE(a.bank_transfer_number, ''), a.bank_closing_balance,
			a.bank_overdraft_contracted_limit, a.bank_overdraft_used_limit, a.bank_automatically_invested_balance,
			COALESCE(a.bank_has_reserved_balance, false),
			COALESCE(a.credit_brand, ''), COALESCE(a.credit_level, ''), a.credit_limit, a.credit_available_limit,
			a.credit_balance_close_date, a.credit_balance_due_date, a.credit_minimum_payment,
			COALESCE(a.credit_status, ''), COALESCE(a.credit_holder_type, ''),
			a.created_at, a.updated_at,
			i.connector_name as bank_name
		FROM pluggy_account a
		JOIN pluggy_item i ON a.item_id = i.id
		WHERE ($1 = '' OR a.type = $1)
		ORDER BY a.name ASC;
	`
	rows, err := d.conn.QueryContext(ctx, query, accountType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DBAccount
	for rows.Next() {
		var acc DBAccount
		err := rows.Scan(
			&acc.ID, &acc.ItemID, &acc.Type, &acc.Subtype, &acc.Name, &acc.MarketingName,
			&acc.Balance, &acc.CurrencyCode, &acc.Number, &acc.Owner, &acc.TaxNumber,
			&acc.BankAgency, &acc.BankTransferNumber, &acc.BankClosingBalance,
			&acc.BankOverdraftContractedLimit, &acc.BankOverdraftUsedLimit, &acc.BankAutomaticallyInvestedBalance,
			&acc.BankHasReservedBalance,
			&acc.CreditBrand, &acc.CreditLevel, &acc.CreditLimit, &acc.CreditAvailableLimit,
			&acc.CreditBalanceCloseDate, &acc.CreditBalanceDueDate, &acc.CreditMinimumPayment,
			&acc.CreditStatus, &acc.CreditHolderType,
			&acc.CreatedAt, &acc.UpdatedAt,
			&acc.BankName,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// ListTransactions busca transações com filtros de conta, tipo, busca e paginação
func (d *DB) ListTransactions(ctx context.Context, accountID, accountType, search string, limit, offset int) ([]DBTransaction, int, error) {
	if limit <= 0 {
		limit = 50
	}
	baseWhere := `
		FROM pluggy_transaction t
		JOIN pluggy_account a ON t.account_id = a.id
		JOIN pluggy_item i ON a.item_id = i.id
		WHERE ($1 = '' OR t.account_id = $1)
		  AND ($2 = '' OR a.type = $2)
		  AND ($3 = '' OR t.description ILIKE '%' || $3 || '%' OR t.category ILIKE '%' || $3 || '%')
	`
	var total int
	countQuery := "SELECT COUNT(*) " + baseWhere
	err := d.conn.QueryRowContext(ctx, countQuery, accountID, accountType, search).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			t.id, t.account_id, t.item_id, t.description, COALESCE(t.description_raw, ''),
			t.amount, t.currency_code, t.date, COALESCE(t.category, 'Geral'), COALESCE(t.category_id, ''),
			COALESCE(t.type, ''), COALESCE(t.status, 'POSTED'), t.balance, COALESCE(t.provider_code, ''),
			COALESCE(t.merchant_name, ''), COALESCE(t.merchant_cnpj, ''),
			t.created_at, t.updated_at,
			a.name as account_name, a.subtype as account_subtype, i.connector_name as bank_name
	` + baseWhere + `
		ORDER BY t.date DESC
		LIMIT $4 OFFSET $5;
	`
	rows, err := d.conn.QueryContext(ctx, query, accountID, accountType, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []DBTransaction
	for rows.Next() {
		var tx DBTransaction
		err := rows.Scan(
			&tx.ID, &tx.AccountID, &tx.ItemID, &tx.Description, &tx.DescriptionRaw,
			&tx.Amount, &tx.CurrencyCode, &tx.Date, &tx.Category, &tx.CategoryID,
			&tx.Type, &tx.Status, &tx.Balance, &tx.ProviderCode,
			&tx.MerchantName, &tx.MerchantCNPJ,
			&tx.CreatedAt, &tx.UpdatedAt,
			&tx.AccountName, &tx.AccountSubtype, &tx.BankName,
		)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListBills busca as faturas de cartões de crédito
func (d *DB) ListBills(ctx context.Context, accountID string) ([]DBCreditCardBill, error) {
	query := `
		SELECT 
			b.id, b.account_id, b.due_date, b.bill_closing_date, b.total_amount, b.currency_code,
			b.minimum_payment_amount, b.allows_installments, b.finance_charges::text, b.payments::text,
			b.created_at, b.updated_at, a.name as card_name
		FROM pluggy_credit_card_bill b
		JOIN pluggy_account a ON b.account_id = a.id
		WHERE ($1 = '' OR b.account_id = $1)
		ORDER BY b.due_date DESC;
	`
	rows, err := d.conn.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DBCreditCardBill
	for rows.Next() {
		var bill DBCreditCardBill
		err := rows.Scan(
			&bill.ID, &bill.AccountID, &bill.DueDate, &bill.BillClosingDate, &bill.TotalAmount, &bill.CurrencyCode,
			&bill.MinimumPaymentAmount, &bill.AllowsInstallments, &bill.FinanceCharges, &bill.Payments,
			&bill.CreatedAt, &bill.UpdatedAt, &bill.CardName,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, bill)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// ListInvestments busca a carteira de investimentos
func (d *DB) ListInvestments(ctx context.Context, itemID string) ([]DBInvestment, error) {
	query := `
		SELECT 
			inv.id, inv.item_id, inv.name, COALESCE(inv.code, ''), COALESCE(inv.isin, ''),
			inv.type, COALESCE(inv.subtype, ''), inv.balance, inv.amount, inv.currency_code,
			inv.due_date, inv.rate, COALESCE(inv.rate_type, ''), inv.fixed_annual_rate,
			inv.last_month_rate, inv.annual_rate, COALESCE(inv.status, 'ACTIVE'),
			COALESCE(inv.institution, ''), inv.created_at, inv.updated_at,
			i.connector_name as institution_name
		FROM pluggy_investment inv
		JOIN pluggy_item i ON inv.item_id = i.id
		WHERE ($1 = '' OR inv.item_id = $1)
		ORDER BY inv.balance DESC;
	`
	rows, err := d.conn.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DBInvestment
	for rows.Next() {
		var inv DBInvestment
		err := rows.Scan(
			&inv.ID, &inv.ItemID, &inv.Name, &inv.Code, &inv.ISIN,
			&inv.Type, &inv.Subtype, &inv.Balance, &inv.Amount, &inv.CurrencyCode,
			&inv.DueDate, &inv.Rate, &inv.RateType, &inv.FixedAnnualRate,
			&inv.LastMonthRate, &inv.AnnualRate, &inv.Status,
			&inv.Institution, &inv.CreatedAt, &inv.UpdatedAt,
			&inv.InstitutionName,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// ListInvestmentTransactions busca movimentações de um investimento
func (d *DB) ListInvestmentTransactions(ctx context.Context, investmentID string) ([]DBInvestmentTransaction, error) {
	query := `
		SELECT 
			it.id, it.investment_id, it.trade_date, it.date, it.type, COALESCE(it.movement_type, ''),
			COALESCE(it.description, ''), it.quantity, it.value, it.amount, it.net_amount,
			it.agreed_rate, COALESCE(it.brokerage_number, ''), it.expenses::text,
			it.created_at, it.updated_at, inv.name as investment_name
		FROM pluggy_investment_transaction it
		JOIN pluggy_investment inv ON it.investment_id = inv.id
		WHERE ($1 = '' OR it.investment_id = $1)
		ORDER BY it.date DESC;
	`
	rows, err := d.conn.QueryContext(ctx, query, investmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DBInvestmentTransaction
	for rows.Next() {
		var it DBInvestmentTransaction
		err := rows.Scan(
			&it.ID, &it.InvestmentID, &it.TradeDate, &it.Date, &it.Type, &it.MovementType,
			&it.Description, &it.Quantity, &it.Value, &it.Amount, &it.NetAmount,
			&it.AgreedRate, &it.BrokerageNumber, &it.Expenses,
			&it.CreatedAt, &it.UpdatedAt, &it.InvestmentName,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// ListWebhookEvents lista os últimos eventos recebidos via webhook
func (d *DB) ListWebhookEvents(ctx context.Context, limit int) ([]DBWebhookEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT 
			id, event, COALESCE(event_id, ''), COALESCE(item_id, ''), COALESCE(account_id, ''),
			COALESCE(triggered_by, ''), payload::text, status, COALESCE(error_message, ''), received_at
		FROM pluggy_webhook_event
		ORDER BY received_at DESC
		LIMIT $1;
	`
	rows, err := d.conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DBWebhookEvent
	for rows.Next() {
		var ev DBWebhookEvent
		err := rows.Scan(
			&ev.ID, &ev.Event, &ev.EventID, &ev.ItemID, &ev.AccountID,
			&ev.TriggeredBy, &ev.Payload, &ev.Status, &ev.ErrorMessage, &ev.ReceivedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

