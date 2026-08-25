package pharmacy

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"pharmacycounter/model"
	"pharmacycounter/persistence"
	"pharmacycounter/report"
	"pharmacycounter/service"
)

type Pharmacy struct {
	Store   *persistence.Store
	Counter *service.Counter
	Path    string
}

func Open(path string) (*Pharmacy, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("path is required")
	}
	store, err := persistence.Open(path)
	if err != nil {
		return nil, err
	}
	counter, err := service.RestoreWithFallback(store)
	if err != nil {
		_ = store.Close()
		return nil, err
	}
	return &Pharmacy{Store: store, Counter: counter, Path: filepath.Clean(path)}, nil
}

func (p *Pharmacy) Close() error {
	if p == nil || p.Store == nil {
		return nil
	}
	return p.Store.Close()
}

func (p *Pharmacy) Register(patientID, name, phone, priority, prescriptionID, medication, dosage string, instructions []string) (report.Receipt, error) {
	patient := model.PatientRecord{ID: patientID, Name: name, Phone: phone, Priority: model.NormalizePriority(priority)}
	order := model.PrescriptionOrder{ID: prescriptionID, PatientID: patientID, Medication: medication, Dosage: dosage, Instructions: append([]string{}, instructions...)}
	result, err := p.Counter.Register(patient, order)
	if err != nil {
		return report.Receipt{}, err
	}
	return report.MakeReceipt(result.Patient, result.Order, result.Ticket), nil
}

func (p *Pharmacy) Call(number, window string) (model.PharmacyTicket, error) {
	return p.Counter.Call(number, window)
}
func (p *Pharmacy) Complete(number, operator string) (model.PharmacyTicket, error) {
	return p.Counter.Complete(number, operator)
}
func (p *Pharmacy) Search(query string) (service.SearchResult, error) {
	return p.Counter.AdvancedSearch(query)
}
func (p *Pharmacy) Summary() string                       { return report.Render(report.Build(p.Counter)) }
func (p *Pharmacy) PendingNumbers() []string              { return numbers(p.Counter.Pending()) }
func (p *Pharmacy) CalledNumbers() []string               { return numbers(p.Counter.Called()) }
func (p *Pharmacy) CompletedNumbers() []string            { return numbers(p.Counter.Completed()) }
func (p *Pharmacy) Dashboard() (service.Dashboard, error) { return p.Counter.Dashboard() }
func (p *Pharmacy) PatientHistory(id string) (service.PatientHistory, error) {
	return p.Counter.History(id)
}
func (p *Pharmacy) ExportState() ([]byte, error) { return p.Counter.ExportState() }
func (p *Pharmacy) Reconcile() error             { return p.Counter.Reconcile() }

func numbers(tickets []model.PharmacyTicket) []string {
	result := make([]string, 0, len(tickets))
	for _, ticket := range tickets {
		result = append(result, ticket.Number)
	}
	return result
}

func (p *Pharmacy) QueueNumbers(status string) []string {
	if p == nil || p.Counter == nil {
		return []string{}
	}
	var tickets []model.PharmacyTicket
	switch status {
	case model.StatusPending:
		tickets = p.Counter.Pending()
	case model.StatusCalled:
		tickets = p.Counter.Called()
	case model.StatusCompleted:
		tickets = p.Counter.Completed()
	default:
		return []string{}
	}
	return model.TicketNumbers(tickets)
}

func (p *Pharmacy) String() string {
	if p == nil {
		return "pharmacy unavailable"
	}
	return fmt.Sprintf("Pharmacy(%s)", p.Path)
}
