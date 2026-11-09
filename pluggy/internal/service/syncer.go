package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pluggy/internal/db"
	"pluggy/internal/logger"
	"pluggy/internal/pluggy"
)

type Syncer struct {
	db     *db.DB
	client *pluggy.Client
	logger *logger.Logger
}

type SyncStats struct {
	ItemsCount                  int `json:"items_count"`
	AccountsCount               int `json:"accounts_count"`
	TransactionsCount           int `json:"transactions_count"`
	CreditCardBillsCount        int `json:"credit_card_bills_count"`
	InvestmentsCount            int `json:"investments_count"`
	InvestmentTransactionsCount int `json:"investment_transactions_count"`
}

func NewSyncer(database *db.DB, client *pluggy.Client, log *logger.Logger) *Syncer {
	return &Syncer{
		db:     database,
		client: client,
		logger: log,
	}
}

// SyncAll sincroniza todos os itens conhecidos no banco de dados
func (s *Syncer) SyncAll(ctx context.Context) (*SyncStats, error) {
	if s.logger != nil {
		s.logger.Info("[Syncer] Iniciando sincronização em lote de todas as instituições...")
	}
	items, err := s.db.ListItems(ctx)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorf("[Syncer] Erro ao listar itens para sincronização: %v", err)
		}
		return nil, fmt.Errorf("erro ao listar itens para sincronização: %w", err)
	}

	totalStats := &SyncStats{}
	for _, item := range items {
		// Chamada externa ativa de forçar refresh no Pluggy desativada temporariamente.
		// _, _ = s.client.SyncItem(ctx, item.ID)

		stats, err := s.SyncItem(ctx, item.ID)
		if err != nil {
			if s.logger != nil {
				s.logger.Errorf("[Syncer] Erro ao sincronizar item %s (%s): %v", item.ID, item.ConnectorName, err)
			}
			continue
		}
		totalStats.ItemsCount++
		totalStats.AccountsCount += stats.AccountsCount
		totalStats.TransactionsCount += stats.TransactionsCount
		totalStats.CreditCardBillsCount += stats.CreditCardBillsCount
		totalStats.InvestmentsCount += stats.InvestmentsCount
		totalStats.InvestmentTransactionsCount += stats.InvestmentTransactionsCount
	}

	if s.logger != nil {
		s.logger.Infof("[Syncer] Sincronização em lote finalizada com sucesso: %+v", totalStats)
	}
	return totalStats, nil
}

// SyncItem baixa e persiste dados completos de uma conexão (contas, extratos, cartões, faturas, investimentos)
func (s *Syncer) SyncItem(ctx context.Context, itemID string) (*SyncStats, error) {
	syncStart := time.Now()
	if s.logger != nil {
		s.logger.Infof("[Syncer] Iniciando sincronização do Item ID: %s", itemID)
	}
	stats := &SyncStats{}

	// 1. Obter dados da instituição e status do item
	item, err := s.client.GetItem(ctx, itemID)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorf("[Syncer] Falha ao consultar item %s na API Pluggy: %v", itemID, err)
		}
		return nil, fmt.Errorf("falha ao consultar item na API Pluggy: %w", err)
	}

	var errorCode, errorMsg string
	if item.Error != nil {
		errorCode = item.Error.Code
		errorMsg = item.Error.Message
	}

	now := time.Now()
	dbItem := &db.DBItem{
		ID:                    item.ID,
		ConnectorID:           item.Connector.ID,
		ConnectorName:         item.Connector.Name,
		ConnectorImageURL:     item.Connector.ImageURL,
		ConnectorPrimaryColor: item.Connector.PrimaryColor,
		Status:                item.Status,
		ExecutionStatus:       item.ExecutionStatus,
		ClientUserID:          item.ClientUserID,
		ErrorCode:             errorCode,
		ErrorMessage:          errorMsg,
		LastUpdatedAt:         &now,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             now,
	}

	if err := s.db.UpsertItem(ctx, dbItem); err != nil {
		if s.logger != nil {
			s.logger.Errorf("[Syncer] Erro ao salvar item %s no banco: %v", itemID, err)
		}
		return nil, fmt.Errorf("erro ao salvar item no banco: %w", err)
	}
	stats.ItemsCount = 1

	// 2. Buscar Contas Correntes, Poupanças e Cartões de Crédito
	accounts, err := s.client.GetAccounts(ctx, itemID)
	if err != nil {
		if s.logger != nil {
			s.logger.Warnf("[Syncer] Aviso ao buscar contas do item %s: %v", itemID, err)
		}
	} else {
		for _, acc := range accounts {
			dbAcc := s.mapAccount(&acc)
			if err := s.db.UpsertAccount(ctx, dbAcc); err != nil {
				if s.logger != nil {
					s.logger.Errorf("[Syncer] Erro ao salvar conta %s: %v", acc.ID, err)
				}
				continue
			}
			stats.AccountsCount++

			// 2.1 Buscar Transações Bancárias ou de Cartão
			txs, err := s.client.GetTransactions(ctx, acc.ID, "", "")
			if err != nil {
				if s.logger != nil {
					s.logger.Errorf("[Syncer] Erro ao buscar transações da conta %s: %v", acc.ID, err)
				}
			} else {
				for _, tx := range txs {
					dbTx := s.mapTransaction(&tx, itemID)
					if err := s.db.UpsertTransaction(ctx, dbTx); err != nil {
						if s.logger != nil {
							s.logger.Errorf("[Syncer] Erro ao salvar transação %s: %v", tx.ID, err)
						}
						continue
					}
					stats.TransactionsCount++
				}
			}

			// 2.2 Se for Cartão de Crédito, buscar Faturas (Bills)
			if acc.Type == "CREDIT" {
				bills, err := s.client.GetBills(ctx, acc.ID)
				if err != nil {
					if s.logger != nil {
						s.logger.Warnf("[Syncer] Aviso ao buscar faturas do cartão %s: %v", acc.ID, err)
					}
				} else {
					for _, bill := range bills {
						dbBill := s.mapCreditCardBill(&bill, acc.ID)
						if err := s.db.UpsertCreditCardBill(ctx, dbBill); err != nil {
							if s.logger != nil {
								s.logger.Errorf("[Syncer] Erro ao salvar fatura %s: %v", bill.ID, err)
							}
							continue
						}
						stats.CreditCardBillsCount++
					}
				}
			}
		}
	}

	// 3. Buscar Carteira de Investimentos
	investments, err := s.client.GetInvestments(ctx, itemID)
	if err != nil {
		if s.logger != nil {
			s.logger.Warnf("[Syncer] Aviso ao buscar investimentos do item %s: %v", itemID, err)
		}
	} else {
		for _, inv := range investments {
			dbInv := s.mapInvestment(&inv, itemID)
			if err := s.db.UpsertInvestment(ctx, dbInv); err != nil {
				if s.logger != nil {
					s.logger.Errorf("[Syncer] Erro ao salvar investimento %s: %v", inv.ID, err)
				}
				continue
			}
			stats.InvestmentsCount++

			// 3.1 Buscar movimentações do investimento
			invTxs, err := s.client.GetInvestmentTransactions(ctx, inv.ID)
			if err != nil {
				if s.logger != nil {
					s.logger.Warnf("[Syncer] Aviso ao buscar transações do investimento %s: %v", inv.ID, err)
				}
			} else {
				for _, it := range invTxs {
					dbIt := s.mapInvestmentTransaction(&it, inv.ID)
					if err := s.db.UpsertInvestmentTransaction(ctx, dbIt); err != nil {
						if s.logger != nil {
							s.logger.Errorf("[Syncer] Erro ao salvar movimentação de investimento %s: %v", it.ID, err)
						}
						continue
					}
					stats.InvestmentTransactionsCount++
				}
			}
		}
	}

	if s.logger != nil {
		s.logger.Infof("[Syncer] Sincronização do Item %s (%s) concluída em %dms: %+v",
			itemID, item.Connector.Name, time.Since(syncStart).Milliseconds(), stats)
	}
	return stats, nil
}

func (s *Syncer) mapAccount(acc *pluggy.Account) *db.DBAccount {
	raw, _ := json.Marshal(acc)
	dbAcc := &db.DBAccount{
		ID:            acc.ID,
		ItemID:        acc.ItemID,
		Type:          acc.Type,
		Subtype:       acc.Subtype,
		Name:          acc.Name,
		MarketingName: acc.MarketingName,
		Balance:       acc.Balance,
		CurrencyCode:  acc.CurrencyCode,
		Number:        acc.Number,
		Owner:         acc.Owner,
		TaxNumber:     acc.TaxNumber,
		RawData:       string(raw),
	}

	if acc.BankData != nil {
		dbAcc.BankTransferNumber = acc.BankData.TransferNumber
		dbAcc.BankClosingBalance = acc.BankData.ClosingBalance
		dbAcc.BankOverdraftContractedLimit = acc.BankData.OverdraftContractedLimit
		dbAcc.BankOverdraftUsedLimit = acc.BankData.OverdraftUsedLimit
		dbAcc.BankAutomaticallyInvestedBalance = acc.BankData.AutomaticallyInvestedBalance
		if acc.BankData.HasReservedBalance != nil {
			dbAcc.BankHasReservedBalance = *acc.BankData.HasReservedBalance
		}
	}

	if acc.CreditData != nil {
		dbAcc.CreditBrand = acc.CreditData.Brand
		dbAcc.CreditLevel = acc.CreditData.Level
		dbAcc.CreditLimit = acc.CreditData.CreditLimit
		dbAcc.CreditAvailableLimit = acc.CreditData.AvailableCreditLimit
		dbAcc.CreditMinimumPayment = acc.CreditData.MinimumPayment
		dbAcc.CreditStatus = acc.CreditData.Status
		dbAcc.CreditHolderType = acc.CreditData.HolderType

		if acc.CreditData.BalanceCloseDate != "" {
			if t, err := time.Parse("2006-01-02", acc.CreditData.BalanceCloseDate); err == nil {
				dbAcc.CreditBalanceCloseDate = &t
			}
		}
		if acc.CreditData.BalanceDueDate != "" {
			if t, err := time.Parse("2006-01-02", acc.CreditData.BalanceDueDate); err == nil {
				dbAcc.CreditBalanceDueDate = &t
			}
		}
	}

	return dbAcc
}

func (s *Syncer) mapTransaction(tx *pluggy.Transaction, itemID string) *db.DBTransaction {
	raw, _ := json.Marshal(tx)
	payData, _ := json.Marshal(tx.PaymentData)

	dbTx := &db.DBTransaction{
		ID:             tx.ID,
		AccountID:      tx.AccountID,
		ItemID:         itemID,
		Description:    tx.Description,
		DescriptionRaw: tx.DescriptionRaw,
		Amount:         tx.Amount,
		CurrencyCode:   tx.CurrencyCode,
		Date:           tx.Date,
		Category:       tx.Category,
		CategoryID:     tx.CategoryID,
		Type:           tx.Type,
		Status:         tx.Status,
		Balance:        tx.Balance,
		ProviderCode:   tx.ProviderCode,
		PaymentData:    string(payData),
		RawData:        string(raw),
	}

	if tx.Merchant != nil {
		dbTx.MerchantName = tx.Merchant.Name
		dbTx.MerchantCNPJ = tx.Merchant.CNPJ
	}

	return dbTx
}

func (s *Syncer) mapCreditCardBill(b *pluggy.CreditCardBill, accountID string) *db.DBCreditCardBill {
	raw, _ := json.Marshal(b)
	charges, _ := json.Marshal(b.FinanceCharges)
	payments, _ := json.Marshal(b.Payments)

	return &db.DBCreditCardBill{
		ID:                   b.ID,
		AccountID:            accountID,
		DueDate:              b.DueDate,
		BillClosingDate:      b.BillClosingDate,
		TotalAmount:          b.TotalAmount,
		CurrencyCode:         b.TotalAmountCurrencyCode,
		MinimumPaymentAmount: b.MinimumPaymentAmount,
		AllowsInstallments:   b.AllowsInstallments,
		FinanceCharges:       string(charges),
		Payments:             string(payments),
		RawData:              string(raw),
	}
}

func (s *Syncer) mapInvestment(inv *pluggy.Investment, itemID string) *db.DBInvestment {
	raw, _ := json.Marshal(inv)

	return &db.DBInvestment{
		ID:              inv.ID,
		ItemID:          itemID,
		Name:            inv.Name,
		Code:            inv.Code,
		ISIN:            inv.ISIN,
		Type:            inv.Type,
		Subtype:         inv.Subtype,
		Balance:         inv.Balance,
		Amount:          inv.Amount,
		CurrencyCode:    inv.CurrencyCode,
		DueDate:         inv.DueDate,
		Rate:            inv.Rate,
		RateType:        inv.RateType,
		FixedAnnualRate: inv.FixedAnnualRate,
		LastMonthRate:   inv.LastMonthRate,
		AnnualRate:      inv.AnnualRate,
		Status:          inv.Status,
		Institution:     inv.Institution,
		RawData:         string(raw),
	}
}

func (s *Syncer) mapInvestmentTransaction(it *pluggy.InvestmentTransaction, investmentID string) *db.DBInvestmentTransaction {
	raw, _ := json.Marshal(it)
	expenses, _ := json.Marshal(it.Expenses)

	return &db.DBInvestmentTransaction{
		ID:              it.ID,
		InvestmentID:    investmentID,
		TradeDate:       it.TradeDate,
		Date:            it.Date,
		Type:            it.Type,
		MovementType:    it.MovementType,
		Description:     it.Description,
		Quantity:        it.Quantity,
		Value:           it.Value,
		Amount:          it.Amount,
		NetAmount:       it.NetAmount,
		AgreedRate:      it.AgreedRate,
		BrokerageNumber: it.BrokerageNumber,
		Expenses:        string(expenses),
		RawData:         string(raw),
	}
}

