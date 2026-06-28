package driven

import (
	"billing/internal/domain"
	"billing/internal/port"
)

// IssuerFile is a concrete implementation of the port.Issuer interface for handling file-based emissions.
type IssuerFile struct {
	filePath string
	pattern  string
	logger   port.Logger
}

// NewIssuerFile creates a new instance of IssuerFile with the specified file path and logger.
func NewIssuerFile(filePath string, filePattern string, logger port.Logger) *IssuerFile {
	return &IssuerFile{
		filePath: filePath,
		pattern:  filePattern,
		logger:   logger,
	}
}

// SendEmission sends the emission data to a file and logs the operation.
func (i *IssuerFile) SendEmission(emission *domain.Emission) error {
	// Implement the logic to write the emission data to a file at i.filePath
	// For example, you could use os.WriteFile or similar methods to write the data.
	i.logger.IPrintf(2, "Sending emission to file: %s", i.filePath)

	// Here, you would serialize the emission data to a suitable format (e.g., JSON, CSV) and write it to the file.
	i.logger.IPrintf(2, "Emission ID: %d, Quantity: %d, Amount: %.2f", emission.ID, emission.Quantity, emission.Amount)
	// Placeholder for actual file writing logic
	return nil
}
