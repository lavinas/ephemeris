package service

import (
	"errors"
	"fmt"

	"planner/internal/domain"
	"planner/internal/port"
)

// resolveInvoiceVendorAndCustomer resolves the vendor nickname and customer nickname for an invoice.
func resolveInvoiceVendorAndCustomer(repo port.Repository, inv *domain.Invoice) (string, string, error) {
	if inv == nil {
		return "", "", errors.New("fatura não fornecida")
	}
	var customer *domain.Customer
	if inv.Plan != nil && inv.Plan.Customer != nil {
		customer = inv.Plan.Customer
	} else if inv.Plan != nil && inv.Plan.CustomerID != 0 && repo != nil {
		c, err := repo.GetCustomerByID(inv.Plan.CustomerID)
		if err != nil {
			return "", "", fmt.Errorf("erro ao buscar cliente %d: %w", inv.Plan.CustomerID, err)
		}
		customer = c
	} else if inv.PlanID != 0 && repo != nil {
		p, err := repo.GetPlanByID(inv.PlanID)
		if err != nil {
			return "", "", fmt.Errorf("erro ao buscar plano %d: %w", inv.PlanID, err)
		}
		if p != nil {
			if p.Customer != nil {
				customer = p.Customer
			} else if p.CustomerID != 0 {
				c, err := repo.GetCustomerByID(p.CustomerID)
				if err != nil {
					return "", "", fmt.Errorf("erro ao buscar cliente %d: %w", p.CustomerID, err)
				}
				customer = c
			}
		}
	}

	if customer == nil {
		return "", "", fmt.Errorf("cliente não encontrado para pré-fatura %d", inv.ID)
	}

	var vendor *domain.Vendor
	if customer.Vendor != nil {
		vendor = customer.Vendor
	} else if customer.VendorID != 0 && repo != nil {
		v, err := repo.GetVendorByID(customer.VendorID)
		if err != nil {
			return "", "", fmt.Errorf("erro ao buscar vendedor %d: %w", customer.VendorID, err)
		}
		vendor = v
	}

	if vendor == nil {
		return "", "", fmt.Errorf("vendedor não encontrado para cliente %s", customer.Nickname)
	}

	return vendor.Nickname, customer.Nickname, nil
}
