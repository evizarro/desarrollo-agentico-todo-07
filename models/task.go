package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// TitleMaxLength es el número máximo de caracteres permitido en el título.
const TitleMaxLength = 200

// DueDateLayout es el formato del día de la fecha límite, el del campo <input type="date">.
const DueDateLayout = "2006-01-02"

// Task representa una tarea de la lista.
type Task struct {
	ID        uuid.UUID `json:"id" db:"id" form:"-"`
	Title     string    `json:"title" db:"title" form:"title"`
	Completed bool      `json:"completed" db:"completed" form:"-"`
	DueDate   time.Time `json:"due_date" db:"due_date" form:"-"`
	CreatedAt time.Time `json:"created_at" db:"created_at" form:"-"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" form:"-"`

	// dueDateInput es la fecha tal como llegó del formulario.
	dueDateInput string `db:"-" form:"-" json:"-"`
}

// Tasks es una colección de Task.
type Tasks []Task

// Today devuelve el día de hoy según el reloj de la aplicación, a medianoche y en UTC.
func Today() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// SetDueDate fija la fecha límite a partir del texto del formulario. Si el texto no es un día
// válido, deja la fecha a cero.
func (t *Task) SetDueDate(raw string) {
	raw = strings.TrimSpace(raw)
	t.dueDateInput = raw
	t.DueDate = time.Time{}
	if parsed, err := time.Parse(DueDateLayout, raw); err == nil {
		t.DueDate = parsed
	}
}

// Validate se ejecuta en cada llamada a los métodos "pop.Validate*".
func (t *Task) Validate(tx *pop.Connection) (*validate.Errors, error) {
	verrs := validate.NewErrors()

	title := strings.TrimSpace(t.Title)
	if title == "" {
		verrs.Add("title", "El título es obligatorio.")
	} else if len([]rune(title)) > TitleMaxLength {
		verrs.Add("title", fmt.Sprintf("El título no puede superar los %d caracteres.", TitleMaxLength))
	}

	if t.DueDate.IsZero() {
		if t.dueDateInput == "" {
			verrs.Add("due_date", "La fecha límite es obligatoria.")
		} else {
			verrs.Add("due_date", "La fecha límite no es una fecha válida.")
		}
	}

	return verrs, nil
}

// ValidateCreate solo se ejecuta al crear: la fecha límite no puede ser anterior a hoy.
func (t *Task) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	verrs := validate.NewErrors()
	if !t.DueDate.IsZero() && t.PastDue() {
		verrs.Add("due_date", "La fecha límite no puede ser anterior a hoy.")
	}
	return verrs, nil
}

// PastDue indica si el día de la fecha límite es anterior a hoy.
func (t Task) PastDue() bool {
	y, m, d := t.DueDate.Date()
	day := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return day.Before(Today())
}

// Overdue indica si la tarea está pendiente y su fecha límite ya pasó.
func (t Task) Overdue() bool {
	return !t.Completed && t.PastDue()
}

// DueDateValue devuelve el valor del campo de fecha del formulario.
func (t Task) DueDateValue() string {
	if t.dueDateInput != "" {
		return t.dueDateInput
	}
	if t.DueDate.IsZero() {
		return ""
	}
	return t.DueDate.Format(DueDateLayout)
}
